import { useEffect, useState } from "react";
import { CopyToClipboard } from "react-copy-to-clipboard";

interface PlayerState {
  status: number;
  name: string;
  artists: string[];
}

export default function App() {
  const [playerState, setPlayerState] = useState<PlayerState>({
    status: 0,
    name: "Connecting.",
    artists: ["Connecting."],
  });

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
        status: 0,
        name: "Connecting.",
        artists: ["Connecting."],
      });
    };

    return socket;
  }

  useEffect(() => {
    let socket = createWebSocket();

    const interval = setInterval(() => {
      if (socket.readyState !== WebSocket.OPEN) {
        socket = createWebSocket();
      }
    }, 2000);

    const interval2 = setInterval(() => {
      if (socket.readyState === WebSocket.OPEN) {
        socket.send(":ping");
      }
    }, 60000);

    return () => {
      clearInterval(interval);
      clearInterval(interval2);
    };
  }, []);

  return (
    <main className="min-h-screen min-w-full flex justify-center items-center">
      <div className="p-8 m-8 inline-block border border-1 border-green-600 rounded-lg whitespace-nowrap overflow-hidden">
        <p className="whitespace-nowrap overflow-hidden text-ellipsis">
          Now playing:
          <CopyToClipboard
            text={playerState.name}
            onCopy={() => {
              alert(`${playerState.name} Copied!`);
            }}
          >
            <span
              className={`font-bold first:ml-4 hover:cursor-pointer ${
                playerState.status !== 200 ? "text-red-600" : "text-green-600"
              }`}
            >
              {playerState.name}
            </span>
          </CopyToClipboard>
        </p>
        <p className="whitespace-nowrap overflow-hidden text-ellipsis">
          Artists:
          {playerState.artists.map((item) => {
            return (
              <CopyToClipboard
                onCopy={() => {
                  alert(`${item} Copied!`);
                }}
                key={item}
                text={item}
              >
                <span
                  className={`font-bold first:ml-4 ml-2 hover:cursor-pointer ${
                    playerState.status !== 200
                      ? "text-red-600"
                      : "text-green-600"
                  }`}
                >
                  {item}
                </span>
              </CopyToClipboard>
            );
          })}
        </p>
      </div>
    </main>
  );
}
