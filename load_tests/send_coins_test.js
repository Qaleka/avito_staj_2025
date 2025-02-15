import http from 'k6/http';
import { check, sleep } from 'k6';
import { login, BASE_URL, RECEIVER_USERNAME, options } from './common.js';

export { options };

export default function () {
    const token = login();
    if (token) {
        const url = `${BASE_URL}/sendCoin`;
        const payload = JSON.stringify({
            toUser: RECEIVER_USERNAME,
            amount: 10,
        });
        const params = {
            headers: {
                'JWT-Token': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
        };
        const res = http.post(url, payload, params);

        if (res.status === 400) {
            console.log(`User ${USERNAME} does not have enough coins to send`);
        } else {
            check(res, {
                'status is 200': (r) => r.status === 200,
            });
        }
    }
    sleep(1);
}