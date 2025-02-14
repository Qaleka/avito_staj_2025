import http from 'k6/http';
import { check, sleep } from 'k6';


export const options = {
    stages: [
        { duration: '15s', target: 1000 },
        { duration: '1m', target: 1000 },
        { duration: '35s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(90)<50'],
        http_req_failed: ['rate<0.0001'],
    },
};

const BASE_URL = 'http://localhost:8080/api';
const USERNAME = `user_${__VU}`;
const PASSWORD = 'test_password';
const ITEM_NAME = 'pen';
const RECEIVER_USERNAME = `user_${__VU}`;

// Функция для аутентификации
function login() {
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

// Функция для получения информации о пользователе
function getUserInfo(token) {
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

// Функция для покупки товара
function buyItem(token) {
    const url = `${BASE_URL}/buy/${ITEM_NAME}`;
    const params = {
        headers: {
            'JWT-Token': `Bearer ${token}`,
        },
    };
    const res = http.get(url, params);

    if (res.status === 400) {
        console.log(`User ${USERNAME} does not have enough coins to buy ${ITEM_NAME}`);
    } else {
        check(res, {
            'status is 200': (r) => r.status === 200,
        });
    }
}

// Функция для передачи монет
function sendCoins(token) {
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

    // Проверяем статус ответа
    if (res.status === 400) {
        console.log(`User ${USERNAME} does not have enough coins to send`);
    } else {
        check(res, {
            'status is 200': (r) => r.status === 200,
        });
    }
}

// Основной сценарий
export default function () {
    const token = login();
    if (token) {
        getUserInfo(token);
        buyItem(token);
        sendCoins(token);
    }
    sleep(1);
}