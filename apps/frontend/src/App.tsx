/* eslint-disable camelcase */
import { useEffect, useState } from "react";

interface PlayerState {
  is_playing: boolean;
  item: Item;
}

interface Item {
  id: string;
  name: string;
  album: {
    images: Image[];
  };
  artists: Artist[];
}

interface Artist {
  id: string;
  name: string;
  images: Image[];
}

interface Image {
  url: string;
}

export default function App() {
  const [isStreaming, setIsStreaming] = useState(false);
  const [playerState, setPlayerState] = useState<PlayerState>({
    is_playing: false,
    item: {
      id: "",
      name: "Connecting.",
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
          name: "Connecting.",
          images: [
            {
              url: "",
            },
          ],
        },
      ],
    },
  });

  function startStream() {
    const endpoint = `${import.meta.env.VITE_BACKEND_ENDPOINT}/api/stream`;
    setIsStreaming(true);

    function errorHandler() {
      setIsStreaming(false);
      setPlayerState({
        is_playing: false,
        item: {
          id: "",
          name: "Connecting.",
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
              name: "Connecting.",
              images: [
                {
                  url: "",
                },
              ],
            },
          ],
        },
      });
    }

    const eventSource = new EventSource(endpoint);

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log(data);
        setPlayerState(data);
      } catch (error) {
        console.error("Failed to parse SSE data", error);
        errorHandler();
      }
    };

    eventSource.onerror = (event) => {
      console.error("SSE error", event);
      errorHandler();
    };

    return eventSource;
  }

  useEffect(() => {
    fetch(`${import.meta.env.VITE_BACKEND_ENDPOINT}/api/state`, {
      method: "GET",
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(response.statusText);
        }
        return response.text();
      })
      .then((resp: string) => {
        const data: PlayerState = JSON.parse(
          resp.substring("data: ".length, resp.length),
        );
        setPlayerState(data);
      })
      .catch(console.error);
    const eventSource = startStream();
    return () => {
      eventSource.close();
      setIsStreaming(false);
    };
  }, []);

  return (
    <main className="min-h-screen min-w-full flex justify-center items-center">
      <div className="p-8 m-8 inline-block border-2 border-green-600 rounded-lg overflow-hidden">
        <p>
          <span>State:</span>
          <span className="m-2 inline-block whitespace-nowrap overflow-hidden text-ellipsis max-w-full align-middle">
            <span
              className={`font-bold ml-2 hover:cursor-pointer ${
                isStreaming === false ? "text-red-600" : "text-green-600"
              }`}
            >
              {isStreaming === false
                ? "Connecting"
                : playerState.is_playing === true
                  ? "Playing"
                  : playerState.is_playing === false && "Paused"}
            </span>
          </span>
        </p>
        <p>
          <span>Name:</span>
          <span className="m-2 inline-block whitespace-nowrap overflow-hidden text-ellipsis max-w-full align-middle">
            {playerState.item.album.images[0].url && (
              <img
                className="inline rounded-full border-2 border-green-600"
                src={playerState.item.album.images[0].url}
                alt={playerState.item.name}
              />
            )}
            <span
              className={`font-bold ml-2 hover:cursor-pointer ${
                isStreaming === false ? "text-red-600" : "text-green-600"
              }`}
            >
              <a
                href={`https://open.spotify.com/track/${playerState.item.id}`}
                target="_blank"
                rel="noreferrer"
              >
                {playerState.item.name}
              </a>
            </span>
          </span>
        </p>
        <p>
          Artists:
          {playerState.item.artists.map((item) => {
            return (
              <span
                key={item.name}
                className="inline-block m-2 whitespace-nowrap overflow-hidden text-ellipsis max-w-full align-middle"
              >
                {item.images[0].url && (
                  <img
                    className="inline rounded-full border-2 border-green-600"
                    src={item.images[0].url}
                    alt={item.name}
                  />
                )}
                <span
                  className={`font-bold ml-2 hover:cursor-pointer ${
                    isStreaming === false ? "text-red-600" : "text-green-600"
                  }`}
                >
                  <a
                    href={`https://open.spotify.com/artist/${item.id}`}
                    target="_blank"
                    rel="noreferrer"
                  >
                    {item.name}
                  </a>
                </span>
              </span>
            );
          })}
        </p>
      </div>
    </main>
  );
}
