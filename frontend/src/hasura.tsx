import {
  ApolloClient,
  gql,
  HttpLink,
  InMemoryCache,
  split,
  useQuery,
  useSubscription,
} from "@apollo/client";
import React, { ReactComponentElement } from "react";
import { getMainDefinition } from "@apollo/client/utilities";
import { WebSocketLink } from "@apollo/client/link/ws";

const httpLink = new HttpLink({
  uri: "http://host.docker.internal:8080/v1/graphql",
});

const wsLink = new WebSocketLink({
  uri: "ws://host.docker.internal:8080/v1/graphql",
  options: {
    reconnect: true,
  },
});

const splitLink = split(
  ({ query }) => {
    const definition = getMainDefinition(query);
    return (
      definition.kind === "OperationDefinition" &&
      definition.operation === "subscription"
    );
  },
  wsLink,
  httpLink
);

export const client = new ApolloClient({
  link: splitLink,
  cache: new InMemoryCache(),
});

const GET_COUNT_QUERY = gql`
  query GetCountStrains {
    cannabis_aggregate {
      aggregate {
        count
      }
    }
  }
`;

const GET_LAST_5_STRAINS_QUERY = gql`
  query GetLast5Strains {
    cannabis(limit: 5, order_by: { id: desc }) {
      strain
    }
  }
`;

const GET_COUNT_SUBSCRIPTION = gql`
  subscription GetCountStrains {
    cannabis_aggregate {
      aggregate {
        count
      }
    }
  }
`;

const GET_LAST_5_STRAINS_SUBSCRIPTION = gql`
  subscription LastStrain {
    cannabis(order_by: { id: desc }, limit: 5) {
      strain
    }
  }
`;

export function GetCountQuery(): string | number | null {
  const { loading, error, data } = useQuery(GET_COUNT_QUERY);

  if (loading) {
    return "Loading...";
  }

  if (error) {
    return `Error: ${error.message}`;
  }

  return data?.cannabis_aggregate?.aggregate?.count || null;
}

export function GetLast5StrainsQuery():
  | ReactComponentElement<any, any>
  | string
  | null {
  const { loading, error, data } = useQuery(GET_LAST_5_STRAINS_QUERY);

  if (loading) {
    return "Loading...";
  }

  if (error) {
    return `Error: ${error.message}`;
  }

  return data.cannabis.map((x: any, i: number) => {
    return <div key={i}>{x.strain}</div>;
  });
}

export function GetCountSubscription(): string | number | null {
  const { data, loading } = useSubscription(GET_COUNT_SUBSCRIPTION);

  if (loading) {
    return "Loading...";
  }

  return data?.cannabis_aggregate?.aggregate?.count || null;
}

export function GetLast5StrainsSubscription():
  | ReactComponentElement<any, any>
  | string
  | null {
  const { data, loading } = useSubscription(GET_LAST_5_STRAINS_SUBSCRIPTION);

  if (loading) {
    return "Loading...";
  }

  return data.cannabis.map((x: any, i: number) => {
    return <div key={i}>{x.strain}</div>;
  });
}
