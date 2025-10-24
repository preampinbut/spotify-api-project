interface PlaybackState {
  is_playing: boolean;
  progress_ms: number;
  item: TrackObject;
}

interface TrackObject {
  id: string;
  name: string;
  artists: Artist[];
  album: Album;
  duration_ms: number;
}

interface Album {
  images: Image[];
}

interface Artist {
  id: string;
  name: string;
  images: Image[];
}

interface Image {
  url: string;
}

interface Artists {
  artists: Artist[];
}
