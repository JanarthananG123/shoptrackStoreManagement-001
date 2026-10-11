import gApiClient from './apiClient';

const gOrderService = {
  fetchOrders() {
    return gApiClient.get('/orders');
  },
  checkoutOrder(pPayload) {
    const lData = pPayload;
    return gApiClient.post('/orders', lData);
  },
  fetchSalesAnalytics(pParams) {
    return gApiClient.get('/analytics/sales', { params: pParams });
  }
};

export default gOrderService;
