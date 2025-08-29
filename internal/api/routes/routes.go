package routes

import (
	"database/sql"
	"net/http"
	"to-do-list/internal/api/handlers"
	"to-do-list/internal/auth"
	"to-do-list/internal/middleware"
	"to-do-list/internal/repository"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(db *sql.DB, tm *auth.TokenManager) http.Handler {
	r := chi.NewRouter()

	userRepo := repository.NewPostgresUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo, tm)

	taskRepo := repository.NewPostgresTaskRepository(db)
	taskHandler := handlers.NewTaskHandler(taskRepo)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Post("/register", userHandler.Register)
			r.Post("/login", userHandler.Login)
		})

		// Все задачи только для авторизованных пользователей
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(tm))

			r.Route("/tasks", func(r chi.Router) {
				r.Post("/", taskHandler.CreateTask)
				r.Get("/", taskHandler.GetUserTasks)
				r.Get("/{taskID}", taskHandler.GetTaskByID)
				r.Patch("/{taskID}", taskHandler.UpdateTask)
				r.Delete("/{taskID}", taskHandler.DeleteTask)
			})
		})
	})

	return r
}
