import gApiClient from './apiClient';

const gOrderService = {
  fetchOrders() {
    return gApiClient.get('/orders');
  },
  checkoutOrder(pPayload) {
    const lData = pPayload;
    return gApiClient.post('/orders', lData);
  }
};

export default gOrderService;
