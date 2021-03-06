package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/stefanprodan/syros/models"
)

func (s *HttpServer) dockerRoutes() chi.Router {
	r := chi.NewRouter()

	// JWT protected
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(s.TokenAuth))
		r.Use(jwtauth.Authenticator)

		r.Get("/hosts", func(w http.ResponseWriter, r *http.Request) {
			hosts, err := s.Repository.AllHosts()
			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				render.PlainText(w, r, err.Error())
				return
			}
			render.JSON(w, r, hosts)
		})

		r.Get("/hosts/{hostID}", func(w http.ResponseWriter, r *http.Request) {
			hostID := chi.URLParam(r, "hostID")

			payload, err := s.Repository.HostContainers(hostID)
			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				render.PlainText(w, r, err.Error())
				return
			}
			render.JSON(w, r, payload)
		})

		r.Get("/environments/{env}", func(w http.ResponseWriter, r *http.Request) {
			env := chi.URLParam(r, "env")

			payload, err := s.Repository.EnvironmentContainers(env)
			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				render.PlainText(w, r, err.Error())
				return
			}

			deployments := models.ChartDto{
				Labels: make([]string, 0),
				Values: make([]int64, 0),
			}

			// aggregate deployments per day based on container created date
			for _, cont := range payload.Containers {
				if cont.State != "running" {
					continue
				}
				date := cont.Created.Format("06-01-02")
				found := -1
				for i, s := range deployments.Labels {
					if s == date {
						found = i
						break
					}
				}
				if found > -1 {
					deployments.Values[found]++
				} else {
					deployments.Labels = append(deployments.Labels, date)
					deployments.Values = append(deployments.Values, 1)
				}
			}

			result := models.EnvironmentDto{
				Host:        payload.Host,
				Containers:  payload.Containers,
				Deployments: deployments,
			}
			render.JSON(w, r, result)
		})

		r.Get("/containers", func(w http.ResponseWriter, r *http.Request) {
			containers, err := s.Repository.AllContainers()
			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				render.PlainText(w, r, err.Error())
				return
			}
			render.JSON(w, r, containers)
		})

		r.Get("/containers/{containerID}", func(w http.ResponseWriter, r *http.Request) {
			containerID := chi.URLParam(r, "containerID")

			payload, err := s.Repository.Container(containerID)
			if err != nil {
				render.Status(r, http.StatusInternalServerError)
				render.PlainText(w, r, err.Error())
				return
			}
			render.JSON(w, r, payload)
		})

	})

	return r
}
