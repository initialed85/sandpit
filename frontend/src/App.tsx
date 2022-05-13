import React from "react";
import "./App.css";
import {
  GetCountQuery,
  GetCountSubscription,
  GetLast5StrainsQuery,
  GetLast5StrainsSubscription,
} from "./hasura";
import { GetMessages } from "./centrifugo";

function App(props: any) {
  return (
    <div className="App">
      <header className="App-header">
        Cannabis
        <div className={"App-outer"}>
          <div className={"App-inner"}>
            <div>Total strains at load...</div>
            <br />
            {GetCountQuery()}
          </div>
          <div className={"App-inner"}>
            <div>Last 5 strains at load...</div>
            <br />
            {GetLast5StrainsQuery()}
          </div>
          <div className={"App-inner"}>
            <div>Known strains right now...</div>
            <br />
            {GetCountSubscription()}
          </div>
          <div className={"App-inner"}>
            <div>Last 5 strains right now...</div>
            <br />
            {GetLast5StrainsSubscription()}
          </div>
        </div>
        <div className={"App-below"}>
          <div>Last 5 statuses right now...</div>
          <br />
          {GetMessages(props.messages)}
        </div>
      </header>
    </div>
  );
}

export default App;
