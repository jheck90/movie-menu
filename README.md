# Movie Menu PWA

A simple Progressive Web App (PWA) for displaying movie posters as a selection menu, perfect for toddlers to pick their movies. Built with Go and the go-app framework.

## Features

- Display movie posters in a grid layout
- Progressive Web App (installable on mobile devices)
- Dark mode by default
- Create and manage multiple movie lists through the UI
- Pull movies from your Radarr server
- Automatic poster fetching and caching from TVDB
- Docker support for easy deployment
- Local file storage (no database required)

## Project Structure

```
movie-menu/
├── main.go              # Main application
├── go.mod              # Go module file
├── go.sum              # Go dependencies
├── Dockerfile          # Docker image definition
├── docker-compose.yml  # Docker compose configuration
├── config/             # Configuration files
│   └── config.json     # API keys and settings
├── lists/              # Movie list JSON files (auto-generated)
├── cache/              # Cached movie posters
└── web/                # Static assets
    └── icon-192.png    # PWA icon
```

## Setup

### 1. Clone and Configure

```bash
# Clone the repository
git clone <your-repo>
cd movie-menu
```

### 2. Configure API Keys

Edit `config/config.json`:

```json
{
  "radarrUrl": "http://your-radarr-server:7878",
  "radarrApiKey": "your-radarr-api-key",
  "tvdbApiKey": "your-tvdb-api-key"
}
```

### 3. Running the Application

#### Local Development

```bash
# Install dependencies
go mod download

# Run the application
go run main.go

# Access at http://localhost:8080
```

#### Using Docker

```bash
# Build and run with Docker Compose
docker-compose up -d

# Or build manually
docker build -t movie-menu .
docker run -p 8080:8080 -v $(pwd)/config:/root/config -v $(pwd)/lists:/root/lists -v $(pwd)/cache:/root/cache movie-menu
```

## Usage

### Creating Movie Lists

1. Open the application in your browser
2. Click "Create New List" button
3. Enter a name for your list (e.g., "Kids Movies", "Disney Favorites")
4. Browse available movies from your Radarr server
5. Select movies to add to your list
6. Save the list

The application will:
- Fetch movie information from Radarr
- Download and cache poster images from TVDB
- Save the list as a JSON file in the `lists/` directory

### Managing Lists

- Switch between lists using the dropdown menu
- Edit existing lists to add or remove movies
- Delete lists you no longer need

### Viewing Movies

- Movie posters are displayed in a responsive grid
- Click on a movie poster to select it
- Images are cached locally for faster loading

## PWA Installation

1. Open the app in a mobile browser
2. You should see an "Add to Home Screen" prompt
3. Install the app for quick access

## Kubernetes Deployment

Create a deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: movie-menu
spec:
  replicas: 1
  selector:
    matchLabels:
      app: movie-menu
  template:
    metadata:
      labels:
        app: movie-menu
    spec:
      containers:
      - name: movie-menu
        image: movie-menu:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /root/config
        - name: lists
          mountPath: /root/lists
        - name: cache
          mountPath: /root/cache
      volumes:
      - name: config
        configMap:
          name: movie-menu-config
      - name: lists
        persistentVolumeClaim:
          claimName: movie-menu-lists
      - name: cache
        persistentVolumeClaim:
          claimName: movie-menu-cache
```

## Technical Details

### API Integration

- **Radarr**: Provides the list of available movies from your media server
- **TVDB**: Supplies high-quality movie poster images

### Caching

Poster images are cached locally to:
- Reduce API calls to TVDB
- Improve loading performance
- Enable offline viewing of previously loaded posters

### Storage

- Configuration: Stored in `config/config.json`
- Movie Lists: Saved as JSON files in `lists/` directory
- Poster Cache: Images stored in `cache/` directory

## Future Enhancements

- Integration with media players (Plex, Jellyfin)
- Parental controls and time restrictions
- Voice selection for non-readers
- Animation transitions
- Admin interface for managing lists
- Automatic list suggestions based on viewing history