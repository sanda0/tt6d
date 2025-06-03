# 📺 TT6D (TodayTVSeries6 Downloader) 🚀

A fun and interactive TV series downloader with a beautiful terminal UI! 🎨

## ✨ Features

- 🎯 Smart link detection and extraction
- 🎬 Support for TV series episodes
- 📦 Support for generic MP4 downloads
- 🖥️ Beautiful terminal UI using [Bubbletea](https://github.com/charmbracelet/bubbletea)
- ⚡ Concurrent downloads with multiple progress bars
- 🎨 Interactive episode selection
- 🎯 No more duplicate downloads
- 📊 Real-time progress tracking
- 🚀 Easy to use!

## 🎮 Usage

```bash
tt6d <webpage_url> <download_folder> [concurrent_downloads]
```

Examples:
```bash
# Download with single progress bar
tt6d https://todaytvseries6.com/series/example /home/user/downloads

# Download with 3 concurrent progress bars
tt6d https://todaytvseries6.com/series/example /home/user/downloads 3
```

## 🎯 Interactive Controls

### Season Selection
- 🔼 Up/Down or j/k: Navigate seasons
- 🎯 Space: Select/deselect season
- ⏩ Enter: Move to episode selection
- ❌ Esc: Quit

### Episode Selection
- 🔼 Up/Down or j/k: Navigate episodes
- 🎯 Space: Select/deselect episode
- 📦 a: Select all episodes
- 🗑️ n: Deselect all
- ⏩ Enter: Start download
- ⬅️ Esc: Back to season selection

## 🌟 Progress Display

Watch your downloads progress with beautiful progress bars:
```
[1/4] episode1.mp4 [████████████░░░░░░░░░░░░] 40% (100/250 MB) 
[2/4] episode2.mp4 [██████████████████░░░░░░] 70% (175/250 MB)
[3/4] episode3.mp4 [████░░░░░░░░░░░░░░░░░░░░] 15% (37/250 MB)
```

## 🚀 Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/tt6d.git

# Build the project
cd tt6d
go build

# Run it!
./tt6d
```

## 📝 License

This project is just for fun! Feel free to use and modify as you like! 🎉

## ⭐ Contributing

Got ideas to make it even more fun? Open an issue or send a PR! 🎨

---
Made with ❤️ and Go
