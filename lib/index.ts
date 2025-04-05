import { UUID } from "node:crypto";
import { writeFileSync } from "node:fs";
import * as tar from "tar";

const clientId = "APP-XRLX4NU88WZZPASK";
const clientSecret = "1755c9a3-bcfd-4e4b-8251-1a2053264db2";
const nilUUID = "00000000-0000-0000-00000000";

type SearchToken = {
  accessToken: UUID;
  refreshToken: UUID;
  scope: string;
  tokenType: string;
  expiresAt: Date;
};

type OrcidRecord = {
  orcidId: UUID;
  name: {
    givenName: string;
    familyName: string;
    creditName?: string;
  };
};

async function getSearchToken(): Promise<[SearchToken, Error | null]> {
  let currentTime = new Date();
  const authResp = await fetch("https://orcid.org/oauth/token", {
    method: "POST",
    headers: { Accept: "application/json", "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      client_id: clientId,
      client_secret: clientSecret,
      grant_type: "client_credentials",
      scope: "/read-public",
    }),
  });
  const data = await authResp.json();
  const searchToken: SearchToken = {
    accessToken: data?.access_token ?? nilUUID,
    scope: data?.scope ?? "",
    refreshToken: data?.refresh_token ?? nilUUID,
    tokenType: data.token_type ?? "",
    expiresAt: new Date(currentTime.getTime() + (data?.expires_in ?? 0) * 1000),
  };
  if (!authResp.ok) {
    const errMsg = data?.error_description;
    return [
      searchToken,
      new Error(`Orcid API err: ${errMsg}`, {
        cause: {
          resp: authResp,
          data: data,
        },
      }),
    ];
  }
  return [searchToken, null];
}

async function getOrcidRecord(orcidId: string) {
  const response = await fetch(`https://pub.orcid.org/v3.0/${orcidId}`, {
    method: "GET",
    headers: {
      Accept: "application/json",
    },
  });
  let record = await response.json();
  return record;
}

class Index {
  private index;
  constructor() {
    this.index = {};
  }

  add(path: string) {
    let idx: any = this.index;
    let steps = path.split("/");
    for (const step of steps) {
      idx[step] = idx[step] || {};
      idx = idx[step];
    }
  }

  write() {
    console.log(JSON.stringify(this.index));
    writeFileSync("index.json", JSON.stringify(this.index));
  }
}

async function buildIndex() {
  const startTime = new Date();
  let tarPath = "/home/adisuper/Downloads/ORCID_2024_10_activities_0.tar.gz";
  console.log("gonna read the file now");
  const index = new Index();
  await tar.t({ file: tarPath, onReadEntry: (entry) => index.add(entry.path) });
  const endTime = new Date();
  const dur = endTime.getTime() - startTime.getTime();
  console.log(`Took ${dur / 1000} seconds`);
  console.log(`Took ${dur / 60000} mins`);
  index.write();
}

Promise.resolve(buildIndex()).then((record) => {
});
