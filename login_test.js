import { sleep } from 'k6';
import { login, options } from './common.js';

export { options };

export default function () {
    const token = login();
    sleep(1);
}