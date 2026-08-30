import http from "k6/http";
import { check, sleep } from "k6";

const baseURL = __ENV.BASE_URL || "http://localhost:8080";

export const options = {
  scenarios: {
    steady_read: {
      executor: "constant-vus",
      vus: 5,
      duration: "30s",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500"],
  },
};

export default function () {
  const response = http.get(`${baseURL}/v1/todos?limit=20&sort=created_at_desc`);
  check(response, { "status is 200": (result) => result.status === 200 });
  sleep(0.2);
}
