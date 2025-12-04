interface Album {
  id: string;
  images: Images[];
}

interface Artist {
  id: string;
  name: string;
  images: Images[];
}

interface Images {
  url: string;
}

interface PlaybackState {
  is_playing: boolean;
  progress_ms: number;
  item: TrackObject;
}

interface ResponseArtist {
  artists: Artist[];
}

interface TrackObject {
  id: string;
  name: string;
  artists: Artist[];
  album: Album;
  duration_ms: number;
}