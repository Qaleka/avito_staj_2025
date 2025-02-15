import http from 'k6/http';
import { check, sleep } from 'k6';
import {login, BASE_URL, ITEM_NAME, options, USERNAME} from './common.js';

export { options };

// Основной сценарий
export default function () {
    const token = login();
    if (token) {
        const url = `${BASE_URL}/buy/${ITEM_NAME}`;
        const params = {
            headers: {
                'JWT-Token': `Bearer ${token}`,
            },
        };
        const res = http.get(url, params);

        // Проверяем статус ответа
        if (res.status === 400) {
            console.log(`User ${USERNAME} does not have enough coins to buy ${ITEM_NAME}`);
        } else {
            check(res, {
                'status is 200': (r) => r.status === 200,
            });
        }
    }
    sleep(1); // Пауза между запросами
}