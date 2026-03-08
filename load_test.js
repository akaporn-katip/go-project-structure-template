import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// 1. Define the load test stages
export const options = {
  stages: [
    { duration: '30s', target: 20 }, // Ramp up to 20 users
    { duration: '1m', target: 20 },  // Stay at 20 users
    { duration: '30s', target: 0 },  // Ramp down to 0
  ],
};

export default function () {
  const url = 'http://localhost:8081/crm-api/v1/customer-profile';
  
  // Generate a unique email for every request to avoid "Unique Constraint" errors
  const uniqueEmail = `user_${randomString(8)}@gmail.com`;

  const payload = JSON.stringify({
    title: "Khun",
    firstname: "Akaporn",
    lastname: "Katip",
    email: uniqueEmail,
    dateOfBirth: "1997-08-18"
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  // 2. Execute the POST request
  const res = http.post(url, payload, params);

  // 3. Validate the response
  check(res, {
    'is status 201 or 200': (r) => r.status === 201 || r.status === 200,
    'transaction time < 500ms': (r) => r.timings.duration < 500,
  });

  // Wait 1 second between iterations per virtual user
  sleep(1);
}
