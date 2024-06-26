/* eslint-disable camelcase */
import dotenv from "dotenv";
dotenv.config();

import fetch from "node-fetch";
import { randomBytes } from "crypto";
import express from "express";
import cors from "cors";

const app = express();
const port = process.env.PORT || 8888;

app.use(cors());

const baseUrl = `${process.env.ENDPOINT}`;
const auth = Buffer.from(
  process.env.CLIENT_ID + ":" + process.env.CLIENT_SECRET,
).toString("base64");

let access_token: string;
let refresh_token: string;
let playerState: any;
let debugResponse: any;
let client: number;

/**
 * This is the only route that should be public
 */

app.get("/api/status", (_req, res) => {
  res.status(200).json(playerState);
});

app.get("/api/stream", (_req, res) => {
  res.setHeader("Cache-Control", "no-cache");
  res.setHeader("Content-Type", "text/event-stream");
  res.setHeader("Access-Control-Allow-Origin", "*");
  res.setHeader("Connection", "keep-alive");
  res.flushHeaders();

  client += 1;
  res.write(JSON.stringify(playerState));
  let interval = setInterval(() => {
    res.write(JSON.stringify(playerState));
  }, 1000 * 3);

  res.on("close", () => {
    client -= 1;
    clearInterval(interval);
    res.end();
  });
});

/**
 * Only restricted the following routes access to some specifig endpoint such as your ip address or local network
 */

app.get("/api/login", (_req, res) => {
  var client_id = process.env.CLIENT_ID;
  var redirect_uri = `${baseUrl}/api/callback`;

  var state = randomBytes(16).toString("hex");
  var scope = "user-read-playback-state";

  res.redirect(
    "https://accounts.spotify.com/authorize?" +
      new URLSearchParams({
        response_type: "code",
        client_id: client_id as string,
        scope: scope,
        redirect_uri: redirect_uri,
        state: state,
      }).toString(),
  );
});

app.get("/api/callback", (req, res) => {
  var code = req.query.code || undefined;
  var state = req.query.state || undefined;

  if (state === undefined) {
    res.redirect(
      "/#" +
        JSON.stringify({
          error: "state_mismatch",
        }),
    );
  } else {
    getAccessToken(code as string);
    res.redirect(302, `${baseUrl}`);
  }
});

/**
 * The following route should be remove or restricted on production
 */

app.get("/api/debug/refresh", (_req, res) => {
  refreshAccessToken();
  res.json({
    status: 200,
  });
});

app.get("/api/debug/response", async (_req, res) => {
  await setPlayerState(true);
  res.json({
    access_token,
    refresh_token,
    ...debugResponse,
  });
});

app.listen(port, () => {
  console.log("Start Interval SetPlayerState");
  setInterval(setPlayerState, 1000 * 3);

  console.log("Start Interval RefreshAccessToken");
  setInterval(refreshAccessToken, 1000 * 60 * 30);
});

function getAccessToken(code: string) {
  let headersList = {
    Authorization: "Basic " + auth,
    "Content-Type": "application/x-www-form-urlencoded",
  };

  let bodyContent = "code=" + code;
  bodyContent += "&redirect_uri=" + `${baseUrl}/api/callback`;
  bodyContent += "&grant_type=authorization_code";

  fetch("https://accounts.spotify.com/api/token", {
    method: "POST",
    body: bodyContent,
    headers: headersList,
  })
    .then((response) => {
      return response.json();
    })
    .then((data: any) => {
      access_token = data.access_token;
      refresh_token = data.refresh_token;
    });
}

async function getPlayingState(): Promise<{}> {
  let headersList = {
    Accept: "application/json",
    "Content-Type": "application/json",
    Authorization: "Bearer " + access_token,
  };

  return await fetch("https://api.spotify.com/v1/me/player", {
    method: "GET",
    headers: headersList,
  })
    .then(async (response) => {
      if (response.status === 204) {
        return {
          is_playing: false,
        };
      }
      return {
        status: response.status,
        ...(await response.json()),
      };
    })
    .then(async (data: any) => {
      debugResponse = data;
      if (data.is_playing === false) {
        return {
          status: 204,
          name: "Does Not Playing Any Track",
          artists: [
            {
              name: process.env.FALLBACK_NAME,
            },
          ],
        };
      }

      let ids = "";

      let index = 0;
      data.item.artists.map((item: any) => {
        ids += item.id;
        if (index < data.item.artists.length - 1) {
          ids += ",";
          index += 1;
        }
      });

      let artists: { name: string; image: string }[] = [];

      await fetch(`https://api.spotify.com/v1/artists?ids=${ids}`, {
        method: "GET",
        headers: headersList,
      })
        .then((response) => response.json())
        .then((datas) => {
          datas.artists.forEach((item: any) => {
            artists.push({
              name: item.name,
              image: item.images[0].url,
            });
          });
        });

      const response = {
        status: 200,
        name: data.item.name,
        image: data.item.album.images[0].url,
        artists: artists,
      };
      return response;
    })
    .catch(async (e) => {
      console.log(e);
      return {
        status: 500,
        name:
          "Something is wrong with the server. What possibly happened is I forgot to login.",
        artists: [
          {
            name: process.env.FALLBACK_NAME,
          },
        ],
      };
    });
}

async function refreshAccessToken() {
  let headersList = {
    Authorization: "Basic " + auth,
    "Content-Type": "application/x-www-form-urlencoded",
  };

  let bodyContent = "grant_type=refresh_token";
  bodyContent += "&refresh_token=" + refresh_token;

  await fetch("https://accounts.spotify.com/api/token", {
    method: "POST",
    body: bodyContent,
    headers: headersList,
  })
    .then((response) => {
      return response.json();
    })
    .then((data: any) => {
      access_token = data.access_token;
      if (data.refresh_token) {
        refresh_token = data.refresh_token;
      }
    });
}

async function setPlayerState(force = false) {
  if (client === 0 && force === false) {
    return;
  }

  let newPlayerState: any = await getPlayingState();

  if (playerState === undefined) {
    playerState = newPlayerState;

    return;
  }

  if (
    playerState.status === newPlayerState.status &&
    playerState.name === newPlayerState.name
  ) {
    return;
  }

  if (
    newPlayerState.status === 400 &&
    playerState.artists[0].name !== process.env.FALLBACK_NAME
  ) {
    playerState = {
      ...playerState,
      status: newPlayerState.status,
    };
  } else {
    playerState = newPlayerState;
  }
}
