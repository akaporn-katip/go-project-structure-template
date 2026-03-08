import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m', target: 20 },
    { duration: '30s', target: 0 },
  ],
  // Optional: Thresholds allow k6 to fail the test if the error rate exceeds a limit
  thresholds: {
    http_req_failed: ['rate < 0.15'], // We expect ~10%, so fail if > 15%
  },
};

export default function () {
  const url = 'http://localhost:8081/crm-api/v1/customer-profile';
  
  // --- 10% Error Logic ---
  // Generate a random number between 0 and 1
  const isErrorRequest = Math.random() < 0.1; 
  
  let email;
  if (isErrorRequest) {
    // Send an invalid email format to trigger a 400/422 error in crm-api
    email = "invalid-email-format-without-at"; 
  } else {
    email = `user_${randomString(8)}@gmail.com`;
  }

  const payload = JSON.stringify({
    title: "Khun",
    firstname: "Akaporn",
    lastname: "Katip",
    email: email,
    dateOfBirth: "1997-08-18"
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post(url, payload, params);

  // --- Adjusted Validation ---
  check(res, {
    // We expect some failures now, so we track both
    'is success (90%)': (r) => r.status === 201 || r.status === 200,
    'is client error (10%)': (r) => r.status >= 400 && r.status < 500,
    'transaction time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}
