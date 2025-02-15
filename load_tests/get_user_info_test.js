import http from 'k6/http';
import { check, sleep } from 'k6';
import { login, BASE_URL, options } from './common.js';

export { options };

export default function () {
    const token = login();
    if (token) {
        const url = `${BASE_URL}/info`;
        const params = {
            headers: {
                'JWT-Token': `Bearer ${token}`,
            },
        };
        const res = http.get(url, params);
        check(res, {
            'status is 200': (r) => r.status === 200,
        });
    }
    sleep(1);
}