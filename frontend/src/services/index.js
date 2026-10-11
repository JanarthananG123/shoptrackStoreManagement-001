import gApiClient from './apiClient';
import gProductService from './productService';
import gOrderService from './orderService';

export const EventService = {
  // Domain modules
  products: gProductService,
  productService: gProductService,
  orders: gOrderService,
  orderService: gOrderService,

  // Direct convenience methods
  fetchProducts() {
    return gProductService.fetchProducts();
  },
  createProduct(pData) {
    return gProductService.createProduct(pData);
  },
  updateStock(pId, pQuantity) {
    return gProductService.updateStock(pId, pQuantity);
  },
  fetchOrders() {
    return gOrderService.fetchOrders();
  },
  checkoutOrder(pPayload) {
    return gOrderService.checkoutOrder(pPayload);
  },
  checkHealth() {
    return gApiClient.get('/getHealth');
  },

  // Underlying configured API client
  apiClient: gApiClient
};

export { gApiClient as apiClient, gProductService as productService, gOrderService as orderService };
export default EventService;
