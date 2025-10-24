import { useEffect, useRef, useState } from "react";

const defaultPlaybackState: PlaybackState = {
  // eslint-disable-next-line camelcase
  is_playing: false,
  // eslint-disable-next-line camelcase
  progress_ms: 0,
  item: {
    id: "",
    name: "Connecting...",
    // eslint-disable-next-line camelcase
    duration_ms: 0,
    album: {
      images: [
        {
          url: "",
        },
      ],
    },
    artists: [
      {
        id: "",
        name: "Connecting...",
        images: [
          {
            url: "",
          },
        ],
      },
    ],
  },
};

const lostConnectionState: PlaybackState = {
  ...defaultPlaybackState,
  item: {
    ...defaultPlaybackState.item,
    name: "Connection Lost",
    artists: [
      {
        ...defaultPlaybackState.item.artists[0],
        name: "Attempting to reconnect...",
      },
    ],
  },
};

export default function usePlaybackState() {
  const [playbackState, setPlaybackState] =
    useState<PlaybackState>(defaultPlaybackState);
  const [isStreaming, setIsStreaming] = useState(false);

  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const progressTimer = useRef<ReturnType<typeof setInterval> | null>(null);

  const endpoint = `${import.meta.env.VITE_BACKEND_ENDPOINT}/api/stream`;
  const stateEndpoint = `${import.meta.env.VITE_BACKEND_ENDPOINT}/api/state`;

  /**
   * Clears the progress timer and sets the UI state to either default (connecting)
   * or lost connection.
   * @param isLost If true, sets the state to show a 'Connection Lost' message.
   */
  function resetState(isLost: boolean = false) {
    if (progressTimer.current) {
      clearInterval(progressTimer.current);
      progressTimer.current = null;
    }
    setPlaybackState(isLost ? lostConnectionState : defaultPlaybackState);
  }

  /**
   * Initializes and manages the Server-Sent Events (SSE) connection.
   */
  function initStream() {
    // Close any existing connection before opening a new one.
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }

    const es = new EventSource(endpoint);
    eventSourceRef.current = es;

    // Handle successful connection establishment.
    es.onopen = () => {
      console.info("SSE connected");
      setIsStreaming(true);
      // Restart the progress timer on a successful connection.
      initProgressTimer();
    };

    // Handle incoming data/state updates from the server.
    es.onmessage = (event) => {
      try {
        const data: PlaybackState = JSON.parse(event.data);
        setPlaybackState(data);
      } catch (err) {
        console.error("Failed to parse SSE data:", err);
      }
    };

    // Handle connection errors (loss of connection).
    es.onerror = (err) => {
      console.error("SSE error: Connection lost", err);
      setIsStreaming(false);

      // Update UI to reflect connection loss.
      resetState(true);

      es.close();

      // Only start one reconnection attempt timer.
      if (!reconnectTimer.current) {
        reconnectTimer.current = setTimeout(() => {
          reconnectTimer.current = null;
          console.warn("Reconnecting to SSE...");
          // Reset UI to 'Connecting...' before retrying the stream.
          setPlaybackState(defaultPlaybackState);
          initStream();
        }, 5000); // 5-second delay before retry
      }
    };
  }

  /**
   * Sets up an interval to manually increment the song progress every second.
   */
  function initProgressTimer() {
    if (progressTimer.current) {
      clearInterval(progressTimer.current);
    }

    progressTimer.current = setInterval(() => {
      setPlaybackState((prev) => {
        // Stop incrementing if the song is paused or if the app is disconnected.
        if (
          !prev.is_playing ||
          prev.item.name === lostConnectionState.item.name
        ) {
          return prev;
        }

        const newProgress = prev.progress_ms + 1000;

        // Cap progress at duration if it exceeds it.
        if (newProgress >= prev.item.duration_ms) {
          return {
            ...prev,
            // eslint-disable-next-line camelcase
            progress_ms: prev.item.duration_ms,
          };
        }

        // Apply the new progress.
        return {
          ...prev,
          // eslint-disable-next-line camelcase
          progress_ms: newProgress,
        };
      });
    }, 1000); // Update progress every 1000ms (1 second)
  }

  // Effect to initialize the progress timer on component mount.
  useEffect(() => {
    initProgressTimer();

    // Cleanup interval on unmount.
    return () => {
      if (progressTimer.current) {
        clearInterval(progressTimer.current);
      }
    };
  }, []);

  // Main effect to fetch initial state and start the SSE stream.
  useEffect(() => {
    /**
     * Fetches the current playback state from a REST endpoint before starting the stream.
     */
    async function fetchInitialState() {
      try {
        const res = await fetch(stateEndpoint);
        if (!res.ok) {
          throw new Error(res.statusText);
        }
        const text = await res.text();
        // The regex removes any 'data: ' prefix if the state endpoint returns SSE-formatted text.
        const data: PlaybackState = JSON.parse(text.replace(/^data:\s*/, ""));
        setPlaybackState(data);
      } catch (err) {
        console.error("Failed to fetch initial state:", err);
      }
    }

    fetchInitialState();
    initStream();

    // Cleanup function runs on unmount.
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
      }
      if (progressTimer.current) {
        clearInterval(progressTimer.current);
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Empty dependency array ensures this runs once on mount.

  return {
    playbackState,
    isStreaming,
  };
}
