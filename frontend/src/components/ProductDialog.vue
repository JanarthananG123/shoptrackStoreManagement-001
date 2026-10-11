<template>
  <v-dialog :value="visible" @input="onDialogInput" max-width="550px" persistent>
    <v-card class="elevation-12 rounded-lg">
      <v-card-title class="headline primary white--text py-3 px-4">
        <v-icon left color="white">mdi-package-variant-plus</v-icon>
        <span>Add New Inventory Item</span>
        <v-spacer></v-spacer>
        <v-btn icon dark small @click="closeDialog">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <v-card-text class="pt-6 pb-2 px-6">
        <v-form ref="form" v-model="valid" lazy-validation @submit.prevent="saveProduct">
          <v-row dense>
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="form.sku"
                label="SKU Code *"
                placeholder="e.g. ELEC-001"
                outlined
                dense
                prepend-inner-icon="mdi-barcode"
                :rules="skuRules"
                required
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="form.name"
                label="Product Name *"
                placeholder="e.g. Ergonomic Keyboard"
                outlined
                dense
                prepend-inner-icon="mdi-format-title"
                :rules="nameRules"
                required
              ></v-text-field>
            </v-col>
          </v-row>

          <v-row dense>
            <v-col cols="12" sm="6">
              <v-text-field
                v-model.number="form.price"
                label="Unit Price (₹) *"
                type="number"
                step="0.01"
                min="0.01"
                outlined
                dense
                prepend-inner-icon="mdi-currency-inr"
                :rules="priceRules"
                required
              ></v-text-field>
            </v-col>

            <v-col cols="12" sm="6">
              <v-text-field
                v-model.number="form.stockQuantity"
                label="Stock Quantity *"
                type="number"
                step="1"
                min="0"
                outlined
                dense
                prepend-inner-icon="mdi-dolly"
                :rules="stockRules"
                required
              ></v-text-field>
            </v-col>
          </v-row>

          <v-alert
            v-if="errorMessage"
            type="error"
            dense
            dismissible
            class="mt-2"
            @dismissed="errorMessage = ''"
          >
            {{ errorMessage }}
          </v-alert>
        </v-form>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions class="px-6 py-3">
        <v-spacer></v-spacer>
        <v-btn text @click="closeDialog" :disabled="loading">
          Cancel
        </v-btn>
        <v-btn
          color="primary"
          class="px-5 font-weight-bold"
          :loading="loading"
          :disabled="!valid || loading"
          @click="saveProduct"
        >
          <v-icon left>mdi-content-save</v-icon>
          Save Product
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
import EventService from '../services';

export default {
  name: 'ProductDialog',
  props: {
    visible: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      valid: true,
      loading: false,
      errorMessage: '',
      form: {
        sku: '',
        name: '',
        price: null,
        stockQuantity: null
      },
      skuRules: [
        v => !!v || 'SKU is required',
        v => (v && v.trim().length > 0) || 'SKU cannot be empty'
      ],
      nameRules: [
        v => !!v || 'Product Name is required',
        v => (v && v.trim().length > 0) || 'Product Name cannot be empty'
      ],
      priceRules: [
        v => v !== null && v !== undefined && v !== '' || 'Price is required',
        v => !isNaN(v) && Number(v) > 0 || 'Price must be greater than 0'
      ],
      stockRules: [
        v => v !== null && v !== undefined && v !== '' || 'Stock Quantity is required',
        v => !isNaN(v) && Number(v) >= 0 || 'Stock Quantity must be 0 or greater',
        v => Number.isInteger(Number(v)) || 'Stock Quantity must be an integer'
      ]
    };
  },
  watch: {
    visible(newVal) {
      if (newVal) {
        this.resetForm();
      }
    }
  },
  methods: {
    onDialogInput(val) {
      if (!val) {
        this.closeDialog();
      }
    },
    resetForm() {
      this.form = {
        sku: '',
        name: '',
        price: null,
        stockQuantity: null
      };
      this.errorMessage = '';
      if (this.$refs.form) {
        this.$refs.form.resetValidation();
      }
    },
    closeDialog() {
      this.$emit('update:visible', false);
      this.$emit('close');
    },
    async saveProduct() {
      if (!this.$refs.form.validate()) {
        return;
      }

      this.loading = true;
      this.errorMessage = '';

      try {
        const payload = {
          sku: this.form.sku.trim(),
          name: this.form.name.trim(),
          price: parseFloat(this.form.price),
          stockQuantity: parseInt(this.form.stockQuantity, 10)
        };

        const response = await EventService.createProduct(payload);
        if (response.data && response.data.success) {
          this.$emit('saved', response.data.data);
          this.closeDialog();
        } else {
          this.errorMessage = response.data.message || 'Failed to save product.';
        }
      } catch (error) {
        if (error.response && error.response.data && error.response.data.message) {
          this.errorMessage = error.response.data.message;
        } else {
          this.errorMessage = 'Network error or backend unavailable.';
        }
      } finally {
        this.loading = false;
      }
    }
  }
};
</script>
