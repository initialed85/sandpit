import Centrifuge from "centrifuge";
import React from "react";

export const centrifuge = new Centrifuge(
  "ws://host.docker.internal:8000/connection/websocket"
);

export const messages: any[] = [];

centrifuge.subscribe("status", function (message) {
  messages.push(message);
});

centrifuge.setToken(
  "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2MzQwMDk2MTksImV4cCI6MTY2NTU0NTYxOSwiYXVkIjoic2FuZHBpdCIsInN1YiI6Im5vYm9keUBsb2NhbGRvbWFpbi5sb2NhbGhvc3QifQ.c7NKmDcUAJuRu-TB_128CAK0hqH9-8Mt65dyyifeJfo"
);

centrifuge.connect();

export function GetMessages(messages: any[]): any {
  const sliced =
    messages && messages.length <= 64 ? messages : messages.slice(-5 - 1, -1);

  const filtered = sliced
    .filter((x: any) => {
      return x?.data?.timestamp && x?.data?.message;
    })
    .map((x: any) => {
      return {
        timestamp: x.data.timestamp,
        message: x.data.message,
      };
    });

  const rows = filtered.reverse().map((x: any, i: number) => {
    return (
      <div className={"App-status-row"} key={i}>
        <div className={"App-status-cell-timestamp"}>{x.timestamp}</div>
        <div className={"App-status-cell-message"}>{x.message}</div>
      </div>
    );
  });

  return <div className={"App-status-container"}>{rows}</div>;
}
