job "movie-menu" {
  datacenters = ["dc1"]
  type        = "service"

  group "web" {
    count = 1

    network {
      port "http" {
        to = 8080
      }
    }

    task "movie-menu" {
      driver = "docker"

      config {
        image = "ghcr.io/jheck90/movie-menu:v1.0.0"
        ports = ["http"]

        # Optional: mount a local config volume (if you're using config.json instead)
        # volumes = ["local/config:/app/config"]
      }

      env {
        RADARR_URL    = "http://192.168.86.3:30143"
        RADARR_API_KEY = "cf911606cb46453a9c4ea9388d8096dd"
        TVDB_API_KEY   = "c70cc66c-8af8-4213-99d6-6397510c6a78"
      }

      resources {
        cpu    = 300
        memory = 256
      }

      service {
        name = "movie-menu"
        port = "http"
        provider = "nomad"
        tags = [
          "traefik.enable=true",
          "traefik.http.routers.menu.rule=Host(`menu.the-casill.net`)",
          "traefik.http.routers.menu.tls.certresolver=letsencrypt",
          "traefik.http.routers.menu.middlewares=localnetwork@file",
        ]
        check {
          name     = "alive"
          type     = "http"
          path     = "/"
          interval = "10s"
          timeout  = "2s"
        }
      }
    }
  }
}
