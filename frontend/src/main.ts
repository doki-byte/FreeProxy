import { createApp } from 'vue'
import { createPinia } from 'pinia';
import ArcoVue from '@arco-design/web-vue';
import ArcoVueIcon from '@arco-design/web-vue/es/icon';
import { Notification } from '@arco-design/web-vue';
import App from '@/App.vue';
import '@arco-design/web-vue/dist/arco.css';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia)
app.use(ArcoVue);
app.use(ArcoVueIcon);
Notification._context = app._context;

app.mount('#app');

