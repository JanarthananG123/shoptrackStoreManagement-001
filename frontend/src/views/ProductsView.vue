<template>
  <v-container fluid>
    <!-- Top Summary Cards -->
    <v-row class="mb-4">
      <v-col cols="12" sm="4">
        <v-card class="mx-auto" elevation="3" outlined>
          <v-list-item two-line>
            <v-list-item-content>
              <div class="text-overline mb-1">Total Products</div>
              <v-list-item-title class="text-h4 font-weight-bold primary--text">
                {{ products.length }}
              </v-list-item-title>
            </v-list-item-content>
            <v-avatar color="primary lighten-4" size="56">
              <v-icon color="primary" large>mdi-cube-outline</v-icon>
            </v-avatar>
          </v-list-item>
        </v-card>
      </v-col>

      <v-col cols="12" sm="4">
        <v-card class="mx-auto" elevation="3" outlined>
          <v-list-item two-line>
            <v-list-item-content>
              <div class="text-overline mb-1">Low Stock Warning</div>
              <v-list-item-title class="text-h4 font-weight-bold warning--text">
                {{ lowStockCount }}
              </v-list-item-title>
            </v-list-item-content>
            <v-avatar color="warning lighten-4" size="56">
              <v-icon color="warning" large>mdi-alert-outline</v-icon>
            </v-avatar>
          </v-list-item>
        </v-card>
      </v-col>

      <v-col cols="12" sm="4">
        <v-card class="mx-auto" elevation="3" outlined>
          <v-list-item two-line>
            <v-list-item-content>
              <div class="text-overline mb-1">Out of Stock</div>
              <v-list-item-title class="text-h4 font-weight-bold error--text">
                {{ outOfStockCount }}
              </v-list-item-title>
            </v-list-item-content>
            <v-avatar color="error lighten-4" size="56">
              <v-icon color="error" large>mdi-package-down</v-icon>
            </v-avatar>
          </v-list-item>
        </v-card>
      </v-col>
    </v-row>

    <!-- Main Data Table Card -->
    <v-card elevation="4" class="rounded-lg">
      <v-card-title class="py-4 px-6">
        <v-icon left color="primary" medium>mdi-format-list-bulleted</v-icon>
        <span class="text-h6 font-weight-bold">Store Inventory</span>
        <v-spacer></v-spacer>

        <!-- Search Input -->
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="Search products..."
          single-line
          hide-details
          outlined
          dense
          class="mr-4"
          style="max-width: 280px;"
        ></v-text-field>

        <!-- Refresh Button -->
        <v-tooltip bottom>
          <template v-slot:activator="{ on, attrs }">
            <v-btn
              icon
              color="primary"
              class="mr-2"
              v-bind="attrs"
              v-on="on"
              :loading="loading"
              @click="fetchProducts"
            >
              <v-icon>mdi-refresh</v-icon>
            </v-btn>
          </template>
          <span>Refresh Data</span>
        </v-tooltip>

        <!-- Add Product Button -->
        <v-btn
          color="primary"
          class="font-weight-bold"
          @click="isDialogVisible = true"
        >
          <v-icon left>mdi-plus</v-icon>
          Add Product
        </v-btn>
      </v-card-title>

      <v-data-table
        :headers="headers"
        :items="products"
        :search="search"
        :loading="loading"
        loading-text="Loading inventory data... Please wait"
        no-data-text="No products available in inventory"
        class="elevation-0"
      >
        <!-- Unit Price Slot -->
        <template v-slot:item.price="{ item }">
          <span class="font-weight-medium">₹{{ Number(item.price).toFixed(2) }}</span>
        </template>

        <!-- Stock Level Slot -->
        <template v-slot:item.stockQuantity="{ item }">
          <span class="font-weight-bold">{{ item.stockQuantity }} units</span>
        </template>

        <!-- Status Chip Slot -->
        <template v-slot:item.status="{ item }">
          <v-chip
            :color="getStatusColor(item.stockQuantity)"
            dark
            small
            class="font-weight-bold"
          >
            <v-icon left x-small>
              {{ getStatusIcon(item.stockQuantity) }}
            </v-icon>
            {{ getStatusText(item.stockQuantity) }}
          </v-chip>
        </template>

        <!-- Actions Slot -->
        <template v-slot:item.actions="{ item }">
          <v-tooltip bottom>
            <template v-slot:activator="{ on, attrs }">
              <v-icon
                small
                color="info"
                class="mr-2"
                v-bind="attrs"
                v-on="on"
                @click="showDetails(item)"
              >
                mdi-eye
              </v-icon>
            </template>
            <span>View Details</span>
          </v-tooltip>
        </template>
      </v-data-table>
    </v-card>

    <!-- Product Creation Modal -->
    <ProductDialog
      :visible.sync="isDialogVisible"
      @saved="onProductSaved"
    />

    <!-- Details View Modal -->
    <v-dialog v-model="isDetailsVisible" max-width="400px">
      <v-card v-if="selectedProduct" class="rounded-lg">
        <v-card-title class="primary white--text py-3">
          <v-icon left color="white">mdi-information-outline</v-icon>
          Product Details
        </v-card-title>
        <v-card-text class="pt-4">
          <v-list dense>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Product ID</v-list-item-subtitle>
                <v-list-item-title class="font-weight-bold">{{ selectedProduct.id }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>SKU Code</v-list-item-subtitle>
                <v-list-item-title class="font-weight-bold">{{ selectedProduct.sku }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Name</v-list-item-subtitle>
                <v-list-item-title class="font-weight-bold">{{ selectedProduct.name }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Unit Price</v-list-item-subtitle>
                <v-list-item-title class="font-weight-bold">₹{{ Number(selectedProduct.price).toFixed(2) }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Stock Level</v-list-item-subtitle>
                <v-list-item-title class="font-weight-bold">{{ selectedProduct.stockQuantity }} units</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-card-text>
        <v-card-actions class="pb-3 px-4">
          <v-spacer></v-spacer>
          <v-btn text color="primary" @click="isDetailsVisible = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Feedback Snackbar -->
    <v-snackbar
      v-model="snackbar.show"
      :color="snackbar.color"
      timeout="4000"
      bottom
      right
    >
      {{ snackbar.text }}
      <template v-slot:action="{ attrs }">
        <v-btn text v-bind="attrs" @click="snackbar.show = false">
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </v-container>
</template>

<script>
import EventService from '../services';
import ProductDialog from '../components/ProductDialog.vue';

export default {
  name: 'ProductsView',
  components: {
    ProductDialog
  },
  data() {
    return {
      search: '',
      loading: false,
      isDialogVisible: false,
      isDetailsVisible: false,
      selectedProduct: null,
      products: [],
      headers: [
        { text: 'SKU', value: 'sku', align: 'start', sortable: true },
        { text: 'Product Name', value: 'name', align: 'start', sortable: true },
        { text: 'Unit Price (₹)', value: 'price', align: 'right', sortable: true },
        { text: 'Stock Level', value: 'stockQuantity', align: 'right', sortable: true },
        { text: 'Stock Status', value: 'status', align: 'center', sortable: false },
        { text: 'Actions', value: 'actions', align: 'center', sortable: false }
      ],
      snackbar: {
        show: false,
        text: '',
        color: 'success'
      }
    };
  },
  computed: {
    lowStockCount() {
      return this.products.filter(p => p.stockQuantity > 0 && p.stockQuantity <= 10).length;
    },
    outOfStockCount() {
      return this.products.filter(p => p.stockQuantity === 0).length;
    }
  },
  mounted() {
    this.fetchProducts();
  },
  methods: {
    async fetchProducts() {
      this.loading = true;
      try {
        const response = await EventService.fetchProducts();
        if (response.data && response.data.success) {
          this.products = response.data.data || [];
          this.$emit('api-status-change', true);
        } else {
          this.showSnackbar('Failed to load products: ' + (response.data.message || 'Unknown error'), 'error');
          this.$emit('api-status-change', false);
        }
      } catch (error) {
        this.showSnackbar('Unable to connect to backend server at http://localhost:8080/api', 'error');
        this.$emit('api-status-change', false);
      } finally {
        this.loading = false;
      }
    },
    getStatusColor(stock) {
      if (stock > 10) return 'green';
      if (stock > 0) return 'orange';
      return 'red';
    },
    getStatusIcon(stock) {
      if (stock > 10) return 'mdi-check-circle';
      if (stock > 0) return 'mdi-alert';
      return 'mdi-close-circle';
    },
    getStatusText(stock) {
      if (stock > 10) return 'In Stock';
      if (stock > 0) return 'Low Stock';
      return 'Out of Stock';
    },
    showDetails(product) {
      this.selectedProduct = product;
      this.isDetailsVisible = true;
    },
    onProductSaved(newProduct) {
      this.showSnackbar(`Product "${newProduct.name}" created successfully!`, 'success');
      this.fetchProducts();
    },
    showSnackbar(text, color = 'success') {
      this.snackbar.text = text;
      this.snackbar.color = color;
      this.snackbar.show = true;
    }
  }
};
</script>
