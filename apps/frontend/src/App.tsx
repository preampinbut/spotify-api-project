/* eslint-disable camelcase */
import { useEffect, useState } from "react";
import { CopyToClipboard } from "react-copy-to-clipboard";

interface PlayerState {
  is_playing: boolean;
  item: Item;
}

interface Item {
  name: string;
  album: {
    images: Image[];
  };
  artists: Artist[];
}

interface Artist {
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
    let endpoint = `${import.meta.env.VITE_BACKEND_ENDPOINT}/api/stream`;

    setIsStreaming(true);

    function errorHandler(err: any) {
      console.error(err);
      setIsStreaming(false);
      setPlayerState({
        is_playing: false,
        item: {
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

    fetch(endpoint)
      .then((response) => {
        const stream = response.body;
        const reader = stream?.getReader();
        function read() {
          reader
            ?.read()
            .then(({ value, done }) => {
              if (done) {
                throw new Error("What is done? How is that possible?");
              }
              const data = new TextDecoder().decode(value);
              console.log(data);
              setPlayerState(JSON.parse(data));
              read();
            })
            .catch((err) => {
              errorHandler(err);
            });
        }
        read();
      })
      .catch((err) => {
        errorHandler(err);
      });
  }

  useEffect(() => {
    fetch(`${import.meta.env.VITE_BACKEND_ENDPOINT}/api/state`, {
      method: "GET",
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(response.statusText);
        }
        return response.json();
      })
      .then((data: PlayerState) => {
        setPlayerState(data);
      })
      .catch(console.error);
    startStream();
  }, []);

  useEffect(() => {
    let interval = setInterval(() => {
      if (!isStreaming) {
        startStream();
      }
    }, 3000);
    return () => {
      clearInterval(interval);
    };
  }, [isStreaming]);

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
                ? "connecting"
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
            <CopyToClipboard
              text={playerState.item.name}
              onCopy={() => {
                alert(`${playerState.item.name} Copied!`);
              }}
            >
              <span
                className={`font-bold ml-2 hover:cursor-pointer ${
                  isStreaming === false ? "text-red-600" : "text-green-600"
                }`}
              >
                {playerState.item.name}
              </span>
            </CopyToClipboard>
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
                <CopyToClipboard
                  onCopy={() => {
                    alert(`${item.name} Copied!`);
                  }}
                  text={item.name}
                >
                  <span
                    className={`font-bold ml-2 hover:cursor-pointer ${
                      isStreaming === false ? "text-red-600" : "text-green-600"
                    }`}
                  >
                    {item.name}
                  </span>
                </CopyToClipboard>
              </span>
            );
          })}
        </p>
      </div>
    </main>
  );
}
