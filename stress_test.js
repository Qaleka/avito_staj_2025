import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '10s', target: 1000 }, // Разгон до 1000 RPS за 10 сек
        { duration: '1m', target: 1000 },  // Держим нагрузку 1 минуту
        { duration: '10s', target: 0 },    // Плавное завершение теста
    ],
    thresholds: {
        http_req_duration: ['p(99.99) < 50'], // 99.99% запросов должны быть быстрее 50 мс
        http_req_failed: ['rate < 0.0001'],   // Процент ошибок < 0.01%
    },
};

export default function () {
    let url = 'http://localhost:8008/api/auth';
    let payload = JSON.stringify({
        username: `user${__VU}`, // Каждый VU отправляет свой логин
        password: 'password123'
    });

    let params = {
        headers: { 'Content-Type': 'application/json' },
    };

    let res = http.post(url, payload, params);

    check(res, {
        'is status 200': (r) => r.status === 200,
        'response time < 50ms': (r) => r.timings.duration < 50,
    });

    sleep(1); // Ждём 1 сек между запросами
}