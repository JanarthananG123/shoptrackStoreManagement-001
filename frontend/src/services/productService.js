import gApiClient from './apiClient';

const gProductService = {
  fetchProducts() {
    return gApiClient.get('/getProducts');
  },
  createProduct(pData) {
    return gApiClient.post('/getProducts', pData);
  },
  updateStock(pId, pQuantity) {
    const lPayload = { stockQuantity: pQuantity };
    return gApiClient.patch(`/getProducts/${pId}`, lPayload);
  }
};

export default gProductService;
