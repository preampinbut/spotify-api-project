interface PlayerState {
  is_playing: boolean;
  item: PlayerStateItem;
}

interface PlayerStateItem {
  id: string;
  name: string;
  artists: PlayerStateItemArtist[];
  album: PlayerStateItemAlbum;
}

interface PlayerStateItemAlbum {
  images: PlayerStateItemImage[];
}

interface PlayerStateItemArtist {
  id: string;
  name: string;
  images: PlayerStateItemImage[];
}

interface PlayerStateItemImage {
  url: string;
}

interface ResponseTypeArtists {
  artists: PlayerStateItemArtist[];
}

