<template>
  <v-container fluid>
    <!-- Top Summary KPI Cards -->
    <v-row class="mb-4">
      <v-col cols="12" sm="6" md="3">
        <v-card class="mx-auto" elevation="3" outlined>
          <v-list-item two-line>
            <v-list-item-content>
              <div class="text-overline mb-1">Total Revenue</div>
              <v-list-item-title class="text-h5 font-weight-bold success--text">
                ₹{{ totalRevenue.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
              </v-list-item-title>
            </v-list-item-content>
            <v-avatar color="success lighten-4" size="52">
              <v-icon color="success" large>mdi-currency-inr</v-icon>
            </v-avatar>
          </v-list-item>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card class="mx-auto" elevation="3" outlined>
          <v-list-item two-line>
            <v-list-item-content>
              <div class="text-overline mb-1">Completed Orders</div>
              <v-list-item-title class="text-h5 font-weight-bold primary--text">
                {{ completedOrdersCount }}
              </v-list-item-title>
            </v-list-item-content>
            <v-avatar color="primary lighten-4" size="52">
              <v-icon color="primary" large>mdi-receipt-text-check</v-icon>
            </v-avatar>
          </v-list-item>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card class="mx-auto" elevation="3" outlined>
          <v-list-item two-line>
            <v-list-item-content>
              <div class="text-overline mb-1">Average Order Value</div>
              <v-list-item-title class="text-h5 font-weight-bold info--text">
                ₹{{ averageOrderValue.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
              </v-list-item-title>
            </v-list-item-content>
            <v-avatar color="info lighten-4" size="52">
              <v-icon color="info" large>mdi-chart-line</v-icon>
            </v-avatar>
          </v-list-item>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card class="mx-auto" elevation="3" outlined>
          <v-list-item two-line>
            <v-list-item-content>
              <div class="text-overline mb-1">Units Sold</div>
              <v-list-item-title class="text-h5 font-weight-bold deep-purple--text">
                {{ totalUnitsSold }}
              </v-list-item-title>
            </v-list-item-content>
            <v-avatar color="deep-purple lighten-4" size="52">
              <v-icon color="deep-purple" large>mdi-package-variant-closed-check</v-icon>
            </v-avatar>
          </v-list-item>
        </v-card>
      </v-col>
    </v-row>

    <!-- Analytics Date Filter & Reporting Toolbar Card -->
    <v-card elevation="2" class="rounded-lg mb-4 pa-4">
      <v-row align="center" dense>
        <!-- Period Filter Button Toggle -->
        <v-col cols="12" md="8" class="d-flex flex-wrap align-center">
          <span class="caption font-weight-bold text-uppercase mr-3 text--secondary">
            <v-icon small left color="primary">mdi-calendar-range</v-icon>
            Period:
          </span>
          <v-btn-toggle
            v-model="selectedPeriod"
            mandatory
            dense
            color="primary"
            class="elevation-0"
            @change="onPeriodChange"
          >
            <v-btn small value="all" class="font-weight-medium text-capitalize">All Time</v-btn>
            <v-btn small value="today" class="font-weight-medium text-capitalize">Today</v-btn>
            <v-btn small value="last7" class="font-weight-medium text-capitalize">Last 7 Days</v-btn>
            <v-btn small value="last30" class="font-weight-medium text-capitalize">Last 30 Days</v-btn>
            <v-btn small value="custom" class="font-weight-medium text-capitalize">Custom Range</v-btn>
          </v-btn-toggle>
        </v-col>

        <!-- Actions: Refresh & CSV Export -->
        <v-col cols="12" md="4" class="d-flex justify-md-end align-center mt-2 mt-md-0">
          <v-btn
            color="success"
            outlined
            dense
            small
            class="font-weight-bold mr-2"
            @click="exportCSV"
          >
            <v-icon left small>mdi-file-delimited</v-icon>
            Export CSV
          </v-btn>
          <v-btn
            color="primary"
            class="font-weight-bold"
            small
            @click="openSalesDialog"
          >
            <v-icon left small>mdi-cart-plus</v-icon>
            Record Sale
          </v-btn>
        </v-col>
      </v-row>

      <!-- Custom Date Range Sub-Bar -->
      <v-expand-transition>
        <div v-if="selectedPeriod === 'custom'" class="pt-3 mt-2">
          <v-divider class="mb-3"></v-divider>
          <v-row dense align="center">
            <v-col cols="12" sm="4" md="3">
              <v-text-field
                v-model="startDate"
                label="Start Date"
                type="date"
                outlined
                dense
                hide-details
                prepend-inner-icon="mdi-calendar-start"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="4" md="3">
              <v-text-field
                v-model="endDate"
                label="End Date"
                type="date"
                outlined
                dense
                hide-details
                prepend-inner-icon="mdi-calendar-end"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="4" md="3" class="d-flex align-center">
              <v-btn
                color="primary"
                small
                class="font-weight-bold mr-2"
                :loading="analyticsLoading"
                @click="applyDateFilter"
              >
                Apply Filter
              </v-btn>
              <v-btn
                text
                small
                @click="clearDateFilter"
              >
                Reset
              </v-btn>
            </v-col>
          </v-row>
        </div>
      </v-expand-transition>
    </v-card>

    <!-- Main Content: Sales History (Left) & Top Products Breakdown (Right) -->
    <v-row>
      <!-- Completed Sales History Table Column -->
      <v-col cols="12" lg="8">
        <v-card elevation="4" class="rounded-lg h-100">
          <v-card-title class="py-4 px-6">
            <v-icon left color="primary" medium>mdi-point-of-sale</v-icon>
            <span class="text-h6 font-weight-bold">Completed Sales History</span>
            <v-spacer></v-spacer>

            <!-- Search Input -->
            <v-text-field
              v-model="search"
              append-icon="mdi-magnify"
              label="Search orders or customers..."
              single-line
              hide-details
              outlined
              dense
              class="mr-2"
              style="max-width: 260px;"
            ></v-text-field>

            <!-- Refresh Button -->
            <v-tooltip bottom>
              <template v-slot:activator="{ on, attrs }">
                <v-btn
                  icon
                  color="primary"
                  v-bind="attrs"
                  v-on="on"
                  :loading="loading"
                  @click="refreshData"
                >
                  <v-icon>mdi-refresh</v-icon>
                </v-btn>
              </template>
              <span>Refresh Sales</span>
            </v-tooltip>
          </v-card-title>

          <v-divider></v-divider>

          <!-- Error Alert -->
          <v-alert
            v-if="errorMessage"
            type="error"
            dismissible
            dense
            class="mx-6 mt-4 mb-0"
            @input="errorMessage = ''"
          >
            {{ errorMessage }}
          </v-alert>

          <!-- Orders Data Table -->
          <v-data-table
            :headers="headers"
            :items="filteredOrders"
            :search="search"
            :loading="loading"
            loading-text="Fetching sales transactions via EventService..."
            no-data-text="No sales records available in selected period"
            class="elevation-0"
          >
            <!-- Order ID Slot -->
            <template v-slot:item.id="{ item }">
              <span class="font-weight-medium primary--text">#{{ item.id }}</span>
            </template>

            <!-- Total Amount Slot -->
            <template v-slot:item.totalAmount="{ item }">
              <span class="font-weight-bold success--text">
                ₹{{ Number(item.totalAmount).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
              </span>
            </template>

            <!-- Items Quantity Slot -->
            <template v-slot:item.itemsCount="{ item }">
              <span>{{ item.itemsCount }} item{{ item.itemsCount === 1 ? '' : 's' }}</span>
            </template>

            <!-- Date Slot -->
            <template v-slot:item.createdAt="{ item }">
              <span>{{ formatDate(item.createdAt) }}</span>
            </template>

            <!-- Status Chip Slot -->
            <template v-slot:item.status="{ item }">
              <v-chip
                small
                color="success lighten-1"
                class="white--text font-weight-bold"
              >
                <v-icon left x-small>mdi-check-circle</v-icon>
                {{ item.status || 'Completed' }}
              </v-chip>
            </template>

            <!-- Actions Slot -->
            <template v-slot:item.actions="{ item }">
              <v-btn
                icon
                small
                color="primary"
                @click="showOrderDetails(item)"
              >
                <v-icon small>mdi-receipt-text</v-icon>
              </v-btn>
            </template>
          </v-data-table>
        </v-card>
      </v-col>

      <!-- Top 5 Best-Selling Products Breakdown Widget -->
      <v-col cols="12" lg="4">
        <v-card elevation="4" class="rounded-lg h-100">
          <v-card-title class="py-4 px-6 d-flex align-center">
            <v-icon left color="amber darken-2" medium>mdi-podium-gold</v-icon>
            <span class="text-h6 font-weight-bold">Top 5 Products</span>
            <v-spacer></v-spacer>
            <v-chip x-small color="primary lighten-4" class="primary--text font-weight-bold">
              Volume
            </v-chip>
          </v-card-title>
          <v-divider></v-divider>

          <v-card-text class="pa-4">
            <div v-if="analyticsLoading" class="text-center py-8">
              <v-progress-circular indeterminate color="primary" size="32"></v-progress-circular>
              <div class="caption mt-2 text--secondary">Calculating sales rankings...</div>
            </div>
            <div v-else-if="topSellingProducts.length === 0" class="text-center py-8 text--secondary">
              <v-icon large color="grey lighten-1">mdi-package-variant</v-icon>
              <div class="mt-2 caption font-weight-medium">No sales recorded in this period</div>
            </div>
            <div v-else>
              <div
                v-for="(lProd, lIdx) in topSellingProducts"
                :key="lProd.product_id"
                class="mb-3 pa-3 rounded-lg"
                :class="$vuetify.theme.dark ? 'grey darken-3' : 'grey lighten-4'"
              >
                <div class="d-flex align-center justify-space-between mb-1">
                  <div class="d-flex align-center">
                    <v-chip
                      x-small
                      :color="lIdx === 0 ? 'amber darken-2' : lIdx === 1 ? 'blue-grey lighten-1' : lIdx === 2 ? 'brown lighten-1' : 'primary'"
                      class="white--text font-weight-bold mr-2"
                    >
                      #{{ lIdx + 1 }}
                    </v-chip>
                    <span class="font-weight-bold text-body-2 text-truncate" style="max-width: 170px;">
                      {{ lProd.name }}
                    </span>
                  </div>
                  <span class="font-weight-bold text-caption success--text">
                    ₹{{ Number(lProd.total_earned).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
                  </span>
                </div>
                <div class="d-flex align-center justify-space-between caption text--secondary mb-1">
                  <span>{{ lProd.units_sold }} unit{{ lProd.units_sold === 1 ? '' : 's' }} sold</span>
                  <span>{{ Math.round((lProd.units_sold / topProductMaxUnits) * 100) }}%</span>
                </div>
                <v-progress-linear
                  :value="(lProd.units_sold / topProductMaxUnits) * 100"
                  height="6"
                  rounded
                  :color="lIdx === 0 ? 'amber darken-2' : 'primary'"
                ></v-progress-linear>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Record Sale Modal Dialog -->
    <v-dialog v-model="lSalesDialog" max-width="720px" persistent>
      <v-card class="rounded-lg">
        <v-card-title class="primary white--text py-3 px-5">
          <v-icon left color="white">mdi-cash-register</v-icon>
          <span>Record New Sale</span>
          <v-spacer></v-spacer>
          <v-btn icon dark small @click="closeSalesDialog">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>

        <v-card-text class="pt-5 px-6">
          <!-- Structured Error Alert -->
          <v-alert
            v-if="formError"
            type="error"
            dense
            outlined
            dismissible
            class="mb-4 font-weight-medium"
            @input="formError = ''"
          >
            {{ formError }}
          </v-alert>

          <!-- Customer Name Input -->
          <v-row dense class="mb-2">
            <v-col cols="12" sm="8">
              <v-text-field
                v-model="lCustomerName"
                label="Customer Name"
                placeholder="Walk-in Customer"
                outlined
                dense
                prepend-inner-icon="mdi-account"
              ></v-text-field>
            </v-col>
          </v-row>

          <div class="text-subtitle-2 font-weight-bold mb-2 primary--text">
            Line Items
          </div>

          <!-- Line Items List -->
          <v-card
            outlined
            class="pa-3 mb-3"
            v-for="(lItem, lIndex) in lLineItems"
            :key="lIndex"
          >
            <v-row dense align="center">
              <!-- Product Autocomplete -->
              <v-col cols="12" sm="5">
                <v-autocomplete
                  v-model="lItem.product"
                  :items="availableProducts"
                  item-text="name"
                  return-object
                  label="Select Product *"
                  placeholder="Type to search..."
                  outlined
                  dense
                  hide-details="auto"
                  prepend-inner-icon="mdi-cube-outline"
                  @change="onProductChange(lItem)"
                >
                  <template v-slot:item="{ item }">
                    <v-list-item-content>
                      <v-list-item-title class="font-weight-medium">{{ item.name }}</v-list-item-title>
                      <v-list-item-subtitle class="caption">
                        SKU: {{ item.sku }} | Price: ₹{{ Number(item.price).toFixed(2) }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                    <v-list-item-action>
                      <v-chip
                        x-small
                        :color="item.stockQuantity > 0 ? 'success' : 'error'"
                        outlined
                        class="font-weight-bold"
                      >
                        {{ item.stockQuantity }} in stock
                      </v-chip>
                    </v-list-item-action>
                  </template>
                </v-autocomplete>
              </v-col>

              <!-- Unit Price (read-only indicator) -->
              <v-col cols="6" sm="2">
                <v-text-field
                  :value="lItem.unitPrice ? '₹' + Number(lItem.unitPrice).toFixed(2) : '-'"
                  label="Unit Price"
                  outlined
                  dense
                  readonly
                  hide-details
                ></v-text-field>
              </v-col>

              <!-- Quantity Input with Stock Validation -->
              <v-col cols="6" sm="2">
                <v-text-field
                  v-model.number="lItem.quantity"
                  type="number"
                  min="1"
                  :max="lItem.maxStock"
                  label="Qty *"
                  outlined
                  dense
                  hide-details="auto"
                  :rules="[
                    v => !!v && v > 0 || '[SVSO001] Min 1',
                    v => !lItem.product || v <= lItem.maxStock || `[SVSO002] Max ${lItem.maxStock}`
                  ]"
                ></v-text-field>
              </v-col>

              <!-- Line Item Total -->
              <v-col cols="8" sm="2" class="text-right">
                <div class="caption text--secondary">Line Total</div>
                <div class="font-weight-bold success--text">
                  ₹{{ ((lItem.quantity || 0) * (lItem.unitPrice || 0)).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
                </div>
              </v-col>

              <!-- Remove Item Button -->
              <v-col cols="4" sm="1" class="text-right">
                <v-btn
                  icon
                  small
                  color="error"
                  :disabled="lLineItems.length <= 1"
                  @click="removeLineItem(lIndex)"
                >
                  <v-icon small>mdi-delete-outline</v-icon>
                </v-btn>
              </v-col>
            </v-row>
          </v-card>

          <!-- Add Line Item Button -->
          <v-btn
            text
            small
            color="primary"
            class="font-weight-bold px-0 mb-3"
            @click="addLineItem"
          >
            <v-icon left small>mdi-plus</v-icon>
            + Add Another Item
          </v-btn>

          <v-divider class="my-3"></v-divider>

          <!-- Grand Total Summary in Footer -->
          <v-row align="center" justify="end" class="px-2">
            <div class="text-h6 font-weight-medium mr-4">Total Amount:</div>
            <div class="text-h5 font-weight-bold success--text">
              ₹{{ lDialogGrandTotal.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
            </div>
          </v-row>
        </v-card-text>

        <v-card-actions class="pb-4 px-6">
          <v-spacer></v-spacer>
          <v-btn text @click="closeSalesDialog" :disabled="submitLoading">
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            class="font-weight-bold px-6"
            :loading="submitLoading"
            @click="submitSale"
          >
            <v-icon left>mdi-check-circle-outline</v-icon>
            Complete Sale
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Order Receipt Dialog -->
    <v-dialog v-model="isDetailsDialogVisible" max-width="480px">
      <v-card v-if="selectedOrder" class="rounded-lg">
        <v-card-title class="primary white--text py-3 px-5">
          <v-icon left color="white">mdi-receipt</v-icon>
          <span>Order Details: #{{ selectedOrder.id }}</span>
          <v-spacer></v-spacer>
          <v-btn icon dark small @click="isDetailsDialogVisible = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text class="pt-4 px-6">
          <v-list dense>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Customer Name</v-list-item-subtitle>
                <v-list-item-title class="font-weight-bold">{{ selectedOrder.customer }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Transaction Date</v-list-item-subtitle>
                <v-list-item-title>{{ formatDate(selectedOrder.createdAt) }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Total Items</v-list-item-subtitle>
                <v-list-item-title class="font-weight-medium">{{ selectedOrder.itemsCount }} units</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Total Amount</v-list-item-subtitle>
                <v-list-item-title class="text-h6 font-weight-bold success--text">
                  ₹{{ Number(selectedOrder.totalAmount).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
                </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Payment Status</v-list-item-subtitle>
                <v-list-item-title>
                  <v-chip small color="success" text-color="white">{{ selectedOrder.status || 'Completed' }}</v-chip>
                </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-card-text>
        <v-card-actions class="pb-3 px-6">
          <v-spacer></v-spacer>
          <v-btn text color="primary" @click="isDetailsDialogVisible = false">Close</v-btn>
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

function formatLocalDate(pDate) {
  const lYear = pDate.getFullYear();
  const lMonth = String(pDate.getMonth() + 1).padStart(2, '0');
  const lDay = String(pDate.getDate()).padStart(2, '0');
  return `${lYear}-${lMonth}-${lDay}`;
}

export default {
  name: 'SalesView',
  data() {
    return {
      search: '',
      loading: false,
      analyticsLoading: false,
      gAnalyticsRequestId: 0,
      serverAnalytics: null,
      selectedPeriod: 'all',
      startDate: '',
      endDate: '',
      appliedStartDate: '',
      appliedEndDate: '',
      orders: [],
      availableProducts: [],
      selectedOrder: null,
      isDetailsDialogVisible: false,
      lSalesDialog: false,
      submitLoading: false,
      formError: '',
      errorMessage: '',
      lCustomerName: 'Walk-in Customer',
      lLineItems: [
        { product: null, quantity: 1, maxStock: 0, unitPrice: 0 }
      ],
      headers: [
        { text: 'Order ID', value: 'id', align: 'start', sortable: true },
        { text: 'Customer', value: 'customer', align: 'start', sortable: true },
        { text: 'Date', value: 'createdAt', align: 'start', sortable: true },
        { text: 'Items', value: 'itemsCount', align: 'right', sortable: true },
        { text: 'Total Amount (₹)', value: 'totalAmount', align: 'right', sortable: true },
        { text: 'Status', value: 'status', align: 'center', sortable: false },
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
    filteredOrders() {
      if (!this.selectedPeriod || this.selectedPeriod === 'all') {
        return this.orders;
      }
      let lStart = null;
      let lEnd = null;
      const lNow = new Date();

      if (this.selectedPeriod === 'today') {
        lStart = new Date(lNow.getFullYear(), lNow.getMonth(), lNow.getDate(), 0, 0, 0, 0);
        lEnd = new Date(lNow.getFullYear(), lNow.getMonth(), lNow.getDate(), 23, 59, 59, 999);
      } else if (this.selectedPeriod === 'last7') {
        lStart = new Date(lNow.getFullYear(), lNow.getMonth(), lNow.getDate() - 7, 0, 0, 0, 0);
        lEnd = new Date(lNow.getFullYear(), lNow.getMonth(), lNow.getDate(), 23, 59, 59, 999);
      } else if (this.selectedPeriod === 'last30') {
        lStart = new Date(lNow.getFullYear(), lNow.getMonth(), lNow.getDate() - 30, 0, 0, 0, 0);
        lEnd = new Date(lNow.getFullYear(), lNow.getMonth(), lNow.getDate(), 23, 59, 59, 999);
      } else if (this.selectedPeriod === 'custom') {
        if (this.appliedStartDate) {
          lStart = new Date(this.appliedStartDate + 'T00:00:00');
        }
        if (this.appliedEndDate) {
          lEnd = new Date(this.appliedEndDate + 'T23:59:59.999');
        }
      }

      return this.orders.filter(pOrder => {
        if (!pOrder.createdAt) return false;
        const lOrderDate = new Date(pOrder.createdAt);
        if (isNaN(lOrderDate.getTime())) return false;
        if (lStart && lOrderDate < lStart) return false;
        if (lEnd && lOrderDate > lEnd) return false;
        return true;
      });
    },
    totalRevenue() {
      if (this.serverAnalytics && typeof this.serverAnalytics.total_revenue === 'number') {
        return this.serverAnalytics.total_revenue;
      }
      return this.filteredOrders.reduce((lSum, lOrder) => lSum + (Number(lOrder.totalAmount) || 0), 0);
    },
    completedOrdersCount() {
      if (this.serverAnalytics && typeof this.serverAnalytics.order_count === 'number') {
        return this.serverAnalytics.order_count;
      }
      return this.filteredOrders.length;
    },
    averageOrderValue() {
      if (this.serverAnalytics && typeof this.serverAnalytics.average_order_value === 'number') {
        return this.serverAnalytics.average_order_value;
      }
      if (this.completedOrdersCount === 0) return 0;
      return this.totalRevenue / this.completedOrdersCount;
    },
    totalUnitsSold() {
      if (this.serverAnalytics && typeof this.serverAnalytics.total_units_sold === 'number') {
        return this.serverAnalytics.total_units_sold;
      }
      return this.filteredOrders.reduce((lSum, lOrder) => lSum + (Number(lOrder.itemsCount) || 0), 0);
    },
    topSellingProducts() {
      if (this.serverAnalytics && Array.isArray(this.serverAnalytics.top_products)) {
        return this.serverAnalytics.top_products;
      }
      // Fallback client-side calculation from filteredOrders
      const lMap = {};
      for (const lOrder of this.filteredOrders) {
        if (Array.isArray(lOrder.items)) {
          for (const lItem of lOrder.items) {
            if (!lMap[lItem.product_id]) {
              const lProd = this.availableProducts.find(p => p.id === lItem.product_id);
              lMap[lItem.product_id] = {
                product_id: lItem.product_id,
                name: lProd ? lProd.name : lItem.product_id,
                units_sold: 0,
                total_earned: 0
              };
            }
            lMap[lItem.product_id].units_sold += Number(lItem.quantity) || 0;
            lMap[lItem.product_id].total_earned += (Number(lItem.quantity) || 0) * (Number(lItem.unit_price) || 0);
          }
        }
      }
      return Object.values(lMap)
        .sort((pA, pB) => pB.units_sold - pA.units_sold)
        .slice(0, 5);
    },
    topProductMaxUnits() {
      if (!this.topSellingProducts || this.topSellingProducts.length === 0) return 1;
      const lMax = Math.max(...this.topSellingProducts.map(p => p.units_sold));
      return lMax > 0 ? lMax : 1;
    },
    lDialogGrandTotal() {
      return this.lLineItems.reduce((lSum, lItem) => {
        const lQty = Number(lItem.quantity) || 0;
        const lPrice = Number(lItem.unitPrice) || 0;
        return lSum + (lQty * lPrice);
      }, 0);
    }
  },
  mounted() {
    this.refreshData();
    this.fetchProducts();
  },
  methods: {
    async refreshData() {
      await Promise.all([
        this.fetchOrders(),
        this.fetchAnalytics()
      ]);
    },
    async fetchOrders() {
      this.loading = true;
      this.errorMessage = '';
      try {
        const lResponse = await EventService.orders.fetchOrders();
        if (lResponse.data && lResponse.data.success) {
          this.orders = lResponse.data.data || [];
          this.$emit('api-status-change', true);
        } else if (Array.isArray(lResponse.data)) {
          this.orders = lResponse.data;
          this.$emit('api-status-change', true);
        } else {
          const lMsg = 'Failed to load orders: ' + (lResponse.data.message || 'Unknown error');
          this.orders = [];
          this.errorMessage = lMsg;
          this.showSnackbar(lMsg, 'error');
        }
      } catch (lError) {
        const lMsg = 'Failed to fetch sales transactions: ' + (lError.response?.data?.message || lError.message || 'Network error');
        this.orders = [];
        this.errorMessage = lMsg;
        this.showSnackbar(lMsg, 'error');
        this.$emit('api-status-change', false);
      } finally {
        this.loading = false;
      }
    },
    async fetchAnalytics() {
      const lRequestId = ++this.gAnalyticsRequestId;
      this.analyticsLoading = true;
      try {
        const lParams = {};
        const lNow = new Date();
        const lTodayStr = formatLocalDate(lNow);

        if (this.selectedPeriod === 'today') {
          lParams.startDate = lTodayStr;
          lParams.endDate = lTodayStr;
        } else if (this.selectedPeriod === 'last7') {
          const lPast7 = new Date(lNow.getFullYear(), lNow.getMonth(), lNow.getDate() - 7);
          lParams.startDate = formatLocalDate(lPast7);
          lParams.endDate = lTodayStr;
        } else if (this.selectedPeriod === 'last30') {
          const lPast30 = new Date(lNow.getFullYear(), lNow.getMonth(), lNow.getDate() - 30);
          lParams.startDate = formatLocalDate(lPast30);
          lParams.endDate = lTodayStr;
        } else if (this.selectedPeriod === 'custom') {
          if (this.appliedStartDate) lParams.startDate = this.appliedStartDate;
          if (this.appliedEndDate) lParams.endDate = this.appliedEndDate;
        }

        const lResponse = await EventService.orders.fetchSalesAnalytics(lParams);
        if (lRequestId !== this.gAnalyticsRequestId) {
          return; // Discard stale in-flight response
        }
        if (lResponse.data && lResponse.data.success && lResponse.data.data) {
          this.serverAnalytics = lResponse.data.data;
        }
      } catch (lError) {
        if (lRequestId !== this.gAnalyticsRequestId) {
          return; // Discard stale in-flight response
        }
        this.serverAnalytics = null;
      } finally {
        if (lRequestId === this.gAnalyticsRequestId) {
          this.analyticsLoading = false;
        }
      }
    },
    onPeriodChange(pPeriod) {
      if (pPeriod !== 'custom') {
        this.startDate = '';
        this.endDate = '';
        this.appliedStartDate = '';
        this.appliedEndDate = '';
        this.fetchAnalytics();
      }
    },
    applyDateFilter() {
      if (this.startDate && this.endDate && this.startDate > this.endDate) {
        this.showSnackbar('[HNSA002] Start date cannot be after end date.', 'error');
        return;
      }
      this.appliedStartDate = this.startDate;
      this.appliedEndDate = this.endDate;
      this.fetchAnalytics();
    },
    clearDateFilter() {
      this.selectedPeriod = 'all';
      this.startDate = '';
      this.endDate = '';
      this.appliedStartDate = '';
      this.appliedEndDate = '';
      this.fetchAnalytics();
    },
    exportCSV() {
      try {
        if (!this.filteredOrders || this.filteredOrders.length === 0) {
          this.showSnackbar('No sales records available to export.', 'warning');
          return;
        }

        const lHeaders = ['Order ID', 'Customer', 'Date', 'Items Count', 'Total Amount (₹)', 'Status'];
        const lRows = this.filteredOrders.map(pOrder => [
          `"${pOrder.id}"`,
          `"${(pOrder.customer || '').replace(/"/g, '""')}"`,
          `"${pOrder.createdAt || ''}"`,
          pOrder.itemsCount || 0,
          (Number(pOrder.totalAmount) || 0).toFixed(2),
          `"${pOrder.status || 'Completed'}"`
        ]);

        const lCsvContent = [lHeaders.join(','), ...lRows.map(pRow => pRow.join(','))].join('\r\n');
        const lBlob = new Blob([lCsvContent], { type: 'text/csv;charset=utf-8;' });
        const lUrl = URL.createObjectURL(lBlob);
        const lLink = document.createElement('a');
        lLink.setAttribute('href', lUrl);
        lLink.setAttribute('download', `shoptrack_sales_${formatLocalDate(new Date())}.csv`);
        document.body.appendChild(lLink);
        lLink.click();
        document.body.removeChild(lLink);
        URL.revokeObjectURL(lUrl);

        this.showSnackbar('Sales report exported to CSV successfully!', 'success');
      } catch (lError) {
        this.showSnackbar('[SVSA001] Failed to export sales CSV: ' + (lError.message || 'Unknown error'), 'error');
      }
    },
    async fetchProducts() {
      try {
        const lResponse = await EventService.products.fetchProducts();
        if (lResponse.data && lResponse.data.success) {
          this.availableProducts = lResponse.data.data || [];
        } else if (Array.isArray(lResponse.data)) {
          this.availableProducts = lResponse.data;
        } else {
          this.availableProducts = [];
        }
      } catch (lError) {
        this.availableProducts = [];
      }
    },
    openSalesDialog() {
      this.formError = '';
      this.lCustomerName = 'Walk-in Customer';
      this.lLineItems = [
        { product: null, quantity: 1, maxStock: 0, unitPrice: 0 }
      ];
      this.fetchProducts();
      this.lSalesDialog = true;
    },
    closeSalesDialog() {
      this.lSalesDialog = false;
      this.formError = '';
    },
    onProductChange(pItem) {
      if (pItem.product) {
        pItem.unitPrice = Number(pItem.product.price) || 0;
        pItem.maxStock = Number(pItem.product.stockQuantity) || 0;
        if (pItem.quantity > pItem.maxStock && pItem.maxStock > 0) {
          pItem.quantity = pItem.maxStock;
        }
      } else {
        pItem.unitPrice = 0;
        pItem.maxStock = 0;
      }
    },
    addLineItem() {
      this.lLineItems.push({
        product: null,
        quantity: 1,
        maxStock: 0,
        unitPrice: 0
      });
    },
    removeLineItem(pIndex) {
      if (this.lLineItems.length > 1) {
        this.lLineItems.splice(pIndex, 1);
      }
    },
    async submitSale() {
      this.formError = '';

      // Validation 1: Empty line items or missing products
      if (!this.lLineItems || this.lLineItems.length === 0) {
        this.formError = '[SVSO001] Order must contain at least one item.';
        return;
      }

      for (let lIndex = 0; lIndex < this.lLineItems.length; lIndex++) {
        const lItem = this.lLineItems[lIndex];
        if (!lItem.product || !lItem.product.id) {
          this.formError = '[SVSO001] Please select a product for all line items.';
          return;
        }
        const lQty = Number(lItem.quantity);
        if (!lQty || lQty <= 0) {
          this.formError = `[SVSO001] Quantity for item #${lIndex + 1} must be at least 1.`;
          return;
        }
        if (lQty > lItem.maxStock) {
          this.formError = `[SVSO002] Quantity for '${lItem.product.name}' (${lQty}) exceeds available stock (${lItem.maxStock}).`;
          return;
        }
      }

      this.submitLoading = true;
      try {
        const lOrderPayload = {
          customer: (this.lCustomerName || 'Walk-in Customer').trim(),
          items: this.lLineItems.map(pItem => ({
            product_id: pItem.product.id,
            quantity: Number(pItem.quantity),
            unit_price: Number(pItem.unitPrice)
          }))
        };

        const lResponse = await EventService.orders.checkoutOrder(lOrderPayload);
        if (lResponse.data && lResponse.data.success) {
          const lCreatedOrder = lResponse.data.data;
          this.showSnackbar(`Sale recorded successfully! Total: ₹${Number(lCreatedOrder.totalAmount).toFixed(2)}`, 'success');
          this.closeSalesDialog();
          this.refreshData();
          this.fetchProducts();
        } else {
          this.formError = lResponse.data.message || '[HNCO004] Failed to record sale.';
        }
      } catch (lError) {
        if (lError.response && lError.response.data && lError.response.data.message) {
          this.formError = lError.response.data.message;
        } else {
          this.formError = '[HNCO001] Network error or backend unavailable.';
        }
      } finally {
        this.submitLoading = false;
      }
    },
    formatDate(pDateStr) {
      if (!pDateStr) return '-';
      const lDate = new Date(pDateStr);
      if (isNaN(lDate.getTime())) return pDateStr;
      return lDate.toLocaleDateString('en-IN', {
        day: '2-digit',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    },
    showOrderDetails(pOrder) {
      this.selectedOrder = pOrder;
      this.isDetailsDialogVisible = true;
    },
    showSnackbar(pText, pColor = 'success') {
      this.snackbar.text = pText;
      this.snackbar.color = pColor;
      this.snackbar.show = true;
    }
  }
};
</script>

<style scoped>
</style>
