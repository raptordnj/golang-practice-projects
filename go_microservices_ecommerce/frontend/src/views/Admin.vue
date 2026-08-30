<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Product {
  id: number
  name: string
  price: number
}

interface Order {
  id: number
  product_id: number
  quantity: number
  total: number
}

const products = ref<Product[]>([])
const orders = ref<Order[]>([])
const loading = ref(true)

const newProductName = ref('')
const newProductPrice = ref('')

const PRODUCT_API = 'http://localhost:8081/products'
const ORDER_API = 'http://localhost:8082/orders'

const fetchProducts = async () => {
  try {
    const res = await fetch(PRODUCT_API)
    products.value = await res.json()
  } catch (e) {
    console.error('Error fetching products', e)
  }
}

const fetchOrders = async () => {
  try {
    const res = await fetch(ORDER_API)
    orders.value = await res.json()
  } catch (e) {
    console.error('Error fetching orders', e)
  }
}

const addProduct = async () => {
  if (!newProductName.value || !newProductPrice.value) return
  
  try {
    const res = await fetch(PRODUCT_API, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: newProductName.value,
        price: parseFloat(newProductPrice.value)
      })
    })
    
    if (res.ok) {
      newProductName.value = ''
      newProductPrice.value = ''
      await fetchProducts()
    }
  } catch (e) {
    console.error('Error adding product', e)
  }
}

const deleteProduct = async (id: number) => {
  if (!confirm('Are you sure you want to delete this product?')) return
  try {
    const res = await fetch(`${PRODUCT_API}/${id}`, {
      method: 'DELETE'
    })
    if (res.ok) {
      await fetchProducts()
    }
  } catch (e) {
    console.error('Error deleting product', e)
  }
}

onMounted(async () => {
  loading.value = true
  await Promise.all([fetchProducts(), fetchOrders()])
  loading.value = false
})
</script>

<template>
  <div class="p-8 max-w-6xl mx-auto space-y-12">
    
    <header class="mb-8">
      <h1 class="text-4xl font-extrabold text-white">Admin Dashboard</h1>
      <p class="text-slate-400 mt-2 text-lg">Manage products and view customer orders.</p>
    </header>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      
      <!-- Manage Products Section -->
      <section class="space-y-6">
        <h2 class="text-2xl font-semibold flex items-center gap-3">
          <span class="w-8 h-8 rounded-lg bg-indigo-500/20 flex items-center justify-center text-indigo-400">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"></path></svg>
          </span>
          Manage Products
        </h2>

        <!-- Add Product Form -->
        <div class="p-6 rounded-2xl bg-slate-800/40 border border-slate-700/50 backdrop-blur-xl space-y-4">
          <h3 class="text-lg font-medium text-slate-300">Add New Product</h3>
          <div class="grid grid-cols-2 gap-4">
            <input v-model="newProductName" type="text" placeholder="Product Name" class="bg-slate-900/50 border border-slate-700 rounded-xl px-4 py-2 text-white focus:outline-none focus:border-indigo-500 transition-colors">
            <input v-model="newProductPrice" type="number" step="0.01" placeholder="Price" class="bg-slate-900/50 border border-slate-700 rounded-xl px-4 py-2 text-white focus:outline-none focus:border-indigo-500 transition-colors">
          </div>
          <button @click="addProduct" class="w-full py-2.5 rounded-xl bg-indigo-500/20 text-indigo-400 font-semibold hover:bg-indigo-500 hover:text-white border border-indigo-500/50 hover:border-indigo-500 transition-all">
            + Add Product
          </button>
        </div>

        <!-- Product List -->
        <div class="p-6 rounded-2xl bg-slate-800/40 border border-slate-700/50 backdrop-blur-xl overflow-hidden flex flex-col h-96">
          <h3 class="text-lg font-medium text-slate-300 mb-4">Product Catalog</h3>
          <div class="flex-1 overflow-y-auto pr-2 custom-scrollbar space-y-3">
            <div v-if="loading" class="text-center text-slate-500 py-4">Loading...</div>
            <div v-else-if="products.length === 0" class="text-center text-slate-500 py-4">No products found.</div>
            
            <div v-else v-for="product in products" :key="product.id" class="flex items-center justify-between p-4 rounded-xl bg-slate-800/80 border border-slate-700 group hover:border-indigo-500/30 transition-colors">
              <div>
                <p class="font-medium text-slate-200">{{ product.name }}</p>
                <p class="text-sm text-indigo-400 font-bold">${{ product.price.toFixed(2) }}</p>
              </div>
              <button @click="deleteProduct(product.id)" class="w-8 h-8 rounded-lg bg-red-500/10 text-red-400 flex items-center justify-center hover:bg-red-500 hover:text-white transition-colors" title="Delete Product">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- All Orders Section -->
      <section class="space-y-6">
        <h2 class="text-2xl font-semibold flex items-center gap-3">
          <span class="w-8 h-8 rounded-lg bg-cyan-500/20 flex items-center justify-center text-cyan-400">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path></svg>
          </span>
          All Orders
        </h2>

        <div class="p-6 rounded-2xl bg-slate-800/40 border border-slate-700/50 backdrop-blur-xl h-[calc(100%-3rem)] overflow-hidden flex flex-col">
          <div v-if="loading" class="space-y-4">
            <div v-for="i in 5" :key="i" class="h-16 rounded-xl bg-slate-700/50 animate-pulse"></div>
          </div>
          
          <div v-else-if="orders.length === 0" class="flex-1 flex flex-col items-center justify-center text-slate-500 space-y-3">
            <svg class="w-12 h-12 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 12H4M8 16l-4-4 4-4"></path></svg>
            <p>No orders placed yet.</p>
          </div>

          <div v-else class="space-y-4 overflow-y-auto pr-2 custom-scrollbar">
            <div 
              v-for="order in orders.slice().reverse()" 
              :key="order.id"
              class="p-4 rounded-xl bg-slate-800/80 border border-slate-700 hover:border-cyan-500/30 transition-colors flex justify-between items-center group"
            >
              <div>
                <p class="text-sm text-slate-400">Order #{{ order.id }}</p>
                <p class="font-medium text-slate-200">Product ID: {{ order.product_id }}</p>
              </div>
              <div class="text-right">
                <p class="text-xs text-slate-500">Qty: {{ order.quantity }}</p>
                <p class="font-bold text-emerald-400">${{ order.total.toFixed(2) }}</p>
              </div>
            </div>
          </div>
        </div>
      </section>

    </div>
  </div>
</template>
