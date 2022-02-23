package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/holedaemon/hole.casa/internal/web/templates"
	"go.uber.org/zap"
)

var ErrMissingOption = errors.New("web: missing required option")

//go:embed static
var static embed.FS

var staticDir fs.FS

func init() {
	var err error
	staticDir, err = fs.Sub(static, "static")
	if err != nil {
		panic(err)
	}
}

type Options struct {
	Addr       string
	Token      string
	GuildID    string
	IgnoreBots bool
}

type Server struct {
	Addr       string
	GuildID    string
	IgnoreBots bool

	logger  *zap.Logger
	discord *discordgo.Session
}

func New(o *Options) (*Server, error) {
	if o.Addr == "" {
		return nil, fmt.Errorf("%w: addr", ErrMissingOption)
	}

	if o.GuildID == "" {
		return nil, fmt.Errorf("%w: guild id", ErrMissingOption)
	}

	if o.Token == "" {
		return nil, fmt.Errorf("%w: token", ErrMissingOption)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	disc, err := discordgo.New("Bot " + o.Token)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		Addr:       o.Addr,
		GuildID:    o.GuildID,
		IgnoreBots: o.IgnoreBots,

		logger:  logger,
		discord: disc,
	}

	return srv, nil
}

// Start instructs the web server to begin listening on addr.
func (s *Server) Start(ctx context.Context) error {
	r := chi.NewMux()
	r.Use(middleware.Recoverer)

	r.Get("/", s.handleIndex)
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.FS(staticDir))))

	srv := &http.Server{
		Addr:    s.Addr,
		Handler: r,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			s.logger.Error("error shutting down server", zap.Error(err))
		}
	}()

	s.logger.Info("listening...", zap.String("addr", s.Addr))
	return srv.ListenAndServe()
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	members, err := s.discord.GuildMembers(s.GuildID, "", 100)
	if err != nil {
		s.logger.Error("error getting members from Discord", zap.Error(err))
	}

	p := make([]*templates.Member, 0, len(members))

	for _, m := range members {
		if s.IgnoreBots && m.User.Bot {
			continue
		}

		p = append(p, &templates.Member{
			Name:      m.User.Username,
			Nick:      m.Nick,
			AvatarURL: m.User.AvatarURL("64x64"),
			Bot:       m.User.Bot,
		})
	}

	templates.WritePageTemplate(w, &templates.IndexPage{
		Members: p,
	})
}
