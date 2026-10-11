import Vue from 'vue';
import VueRouter from 'vue-router';
import ProductsView from '../views/ProductsView.vue';
import SalesView from '../views/SalesView.vue';

Vue.use(VueRouter);

const routes = [
  {
    path: '/',
    name: 'Products',
    component: ProductsView
  },
  {
    path: '/products',
    redirect: '/'
  },
  {
    path: '/sales',
    name: 'Sales',
    component: SalesView
  }
];

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
});

export default router;
