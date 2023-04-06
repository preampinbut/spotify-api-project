import { useEffect, useRef, useState } from "react";
import { CopyToClipboard } from "react-copy-to-clipboard";

interface PlayerState {
  status: number;
  name: string;
  image?: string;
  artists: Artist[];
}

interface Artist {
  name: string;
  image?: string;
}

export default function App() {
  const [playerState, setPlayerState] = useState<PlayerState>({
    status: 500,
    name: "Connecting.",
    artists: [
      {
        name: "Connecting.",
      },
    ],
  });
  let ws = useRef<WebSocket>();

  function createWebSocket() {
    let socket = new WebSocket(
      `${import.meta.env.VITE_BACKEND_WEBSOCKET}/api/websocket`
    );

    socket.onopen = function () {
      console.log("[open]");
    };

    socket.onmessage = function (event) {
      let data = event.data as string;

      let command = data.split(" ")[0];
      if (command === ":pong") {
        return;
      }

      console.log(`[message] ${event.data}`);
      setPlayerState(JSON.parse(event.data));
    };

    socket.onerror = function () {
      console.log("[error]");
    };

    socket.onclose = function () {
      console.log("[close]");
      setPlayerState({
        status: 500,
        name: "Connecting.",
        artists: [
          {
            name: "Connecting.",
          },
        ],
      });
    };

    return socket;
  }

  useEffect(() => {
    if (ws.current) {
      ws.current.onmessage = function (event) {
        let data = event.data as string;

        let command = data.split(" ")[0];
        if (command === ":pong") {
          return;
        }

        console.log(`[message] ${event.data}`);

        let message: PlayerState = JSON.parse(event.data);

        if (
          message.status === 400 &&
          playerState.artists[0].name !== "Pream Pinbut"
        ) {
          return setPlayerState({
            ...playerState,
            status: message.status,
          });
        }

        setPlayerState(message);
      };
    }
  }, [playerState]);

  useEffect(() => {
    ws.current = createWebSocket();

    const interval = setInterval(() => {
      if (ws.current!.readyState !== WebSocket.OPEN) {
        ws.current = createWebSocket();
      }
    }, 2000);

    const interval2 = setInterval(() => {
      if (ws.current!.readyState === WebSocket.OPEN) {
        ws.current!.send(":ping");
      }
    }, 60000);

    return () => {
      clearInterval(interval);
      clearInterval(interval2);
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
                playerState.status === 500 ? "text-red-600" : "text-green-600"
              }`}
            >
              {playerState.status === 200 ? "Playing" : "Paused"}
            </span>
          </span>
        </p>
        <p>
          <span>Now playing:</span>
          <span className="m-2 inline-block whitespace-nowrap overflow-hidden text-ellipsis max-w-full align-middle">
            {playerState.image && (
              <img
                className="inline rounded-full border-2 border-green-600"
                src={playerState.image}
                alt={playerState.name}
              />
            )}
            <CopyToClipboard
              text={playerState.name}
              onCopy={() => {
                alert(`${playerState.name} Copied!`);
              }}
            >
              <span
                className={`font-bold ml-2 hover:cursor-pointer ${
                  playerState.status === 500 ? "text-red-600" : "text-green-600"
                }`}
              >
                {playerState.name}
              </span>
            </CopyToClipboard>
          </span>
        </p>
        <p>
          Artists:
          {playerState.artists.map((item) => {
            return (
              <span
                key={item.name}
                className="inline-block m-2 whitespace-nowrap overflow-hidden text-ellipsis max-w-full align-middle"
              >
                {item.image && (
                  <img
                    className="inline rounded-full border-2 border-green-600"
                    src={item.image}
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
                      playerState.status === 500
                        ? "text-red-600"
                        : "text-green-600"
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
