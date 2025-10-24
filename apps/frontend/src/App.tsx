import usePlaybackState from "./usePlaybackState";
import { formatTime } from "./utils";

export default function App() {
  const { playbackState } = usePlaybackState();

  // --- Playback Indicator Logic ---
  const indicator = playbackState.is_playing ? (
    // State: Playing
    <span className="ml-0 text-primary-1 text-sm flex items-center">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        className="h-4 w-4 mr-1"
        viewBox="0 0 20 20"
        fill="currentColor"
      >
        {/* SVG path for the Play icon (filled circle with a triangle inside) */}
        <path
          fillRule="evenodd"
          d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z"
          clipRule="evenodd"
        />
      </svg>
      Playing
    </span>
  ) : (
    // State: Paused
    <span className="ml-0 text-primary-1 text-sm flex items-center">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        className="h-4 w-4 mr-1"
        viewBox="0 0 20 20"
        fill="currentColor"
      >
        {/* SVG path for the Pause icon (filled circle with two bars inside) */}
        <path
          fillRule="evenodd"
          d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zM7 8a1 1 0 012 0v4a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v4a1 1 0 102 0V8a1 1 0 00-1-1z"
          clipRule="evenodd"
        />
      </svg>
      Paused
    </span>
  );

  const albumImageUrl = playbackState.item.album.images?.[0]?.url;

  // --- Album Cover Display Logic ---
  const albumCover = albumImageUrl ? (
    <img
      src={albumImageUrl}
      className="w-full h-full object-cover"
    />
  ) : (
    // Placeholder div when album art is unavailable.
    <div className="w-full h-full bg-secondary-3 flex items-center justify-center text-gray-500 text-xs">
      No Album Art
    </div>
  );

  return (
    <main className="min-h-screen flex justify-center items-center bg-secondary-1">
      <div className="md:w-3xl p-6 bg-secondary-2 rounded-2xl shadow-lg">
        <div className="flex flex-col md:flex-row gap-6 items-center w-full">
          {/* Album Cover */}
          <div className="w-full md:w-1/2 aspect-square overflow-hidden rounded-xl">
            {albumCover}
          </div>

          {/* Song Info */}
          <div className="flex flex-col justify-center text-center md:text-left gap-2 w-full md:w-1/2 overflow-hidden">
            <p className="text-xl font-semibold truncate">
              {playbackState.item.name}
            </p>
            <p className="text-md text-gray-300 truncate">
              {playbackState.item.artists.map((a) => a.name).join(", ")}
            </p>

            {/* Playing State */}
            {indicator}

            {/* Progress Bar */}
            <div className="mt-4">
              <div className="w-full h-1 bg-gray-700 rounded-full">
                <div
                  className="h-1 bg-primary-1 rounded-full transition-all duration-300"
                  style={{
                    width: `${
                      (playbackState.progress_ms /
                        playbackState.item.duration_ms) *
                      100
                    }%`,
                  }}
                ></div>
              </div>
              <p className="text-sm text-gray-400 mt-1">
                {formatTime(playbackState.progress_ms)} /
                {formatTime(playbackState.item.duration_ms)}
              </p>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
