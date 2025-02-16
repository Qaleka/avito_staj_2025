import http from 'k6/http';
import { check } from 'k6';

export const BASE_URL = 'http://localhost:8080/api';
export const USERNAME = `user_${__VU}`;
export const PASSWORD = 'test_password';
export const ITEM_NAME = 'pen';
export const RECEIVER_USERNAME = `user_${__VU}`;

export const options = {
    stages: [
        { duration: '30s', target: 1000 },
        { duration: '1m', target: 1000 },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(90)<50'],
        http_req_failed: ['rate<0.0001'],
    },
};

export function login() {
    const url = `${BASE_URL}/auth`;
    const payload = JSON.stringify({
        username: USERNAME,
        password: PASSWORD,
    });
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };
    const res = http.post(url, payload, params);
    check(res, {
        'status is 200': (r) => r.status === 200,
        'token received': (r) => r.json('token') !== null,
    });
    return res.json('token');
}