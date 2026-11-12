import { CatalogBatch, Product, ProductListResponse } from '../types';

const BASE_URL = '/api';

export async function fetchHealth() {
  const res = await fetch(`${BASE_URL}/health`);
  return res.json();
}

export async function fetchStats() {
  const res = await fetch(`${BASE_URL}/stats`);
  return res.json();
}

export async function fetchBatches(): Promise<CatalogBatch[]> {
  const res = await fetch(`${BASE_URL}/batches`);
  if (!res.ok) throw new Error('Failed to fetch batches');
  return res.json();
}

export async function fetchProducts(
  batchId: string,
  params: { status?: string; compliance?: string; search?: string; limit?: number; offset?: number } = {}
): Promise<ProductListResponse> {
  const query = new URLSearchParams();
  if (params.status) query.set('status', params.status);
  if (params.compliance) query.set('compliance', params.compliance);
  if (params.search) query.set('search', params.search);
  if (params.limit) query.set('limit', params.limit.toString());
  if (params.offset) query.set('offset', params.offset.toString());

  const res = await fetch(`${BASE_URL}/batches/${batchId}/products?${query.toString()}`);
  if (!res.ok) throw new Error('Failed to fetch products');
  return res.json();
}

export async function updateProduct(id: string, updates: Partial<Product>): Promise<Product> {
  const res = await fetch(`${BASE_URL}/products/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(updates),
  });
  if (!res.ok) throw new Error('Failed to update product');
  return res.json();
}

export async function approveAllBatchProducts(batchId: string): Promise<{ approved_count: number }> {
  const res = await fetch(`${BASE_URL}/batches/${batchId}/approve-all`, {
    method: 'POST',
  });
  if (!res.ok) throw new Error('Failed to approve batch');
  return res.json();
}

export async function uploadFeed(file: File, name: string): Promise<{ batch_id: string }> {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('name', name);

  const res = await fetch(`${BASE_URL}/batches/upload`, {
    method: 'POST',
    body: formData,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Upload failed' }));
    throw new Error(err.error || 'Upload failed');
  }
  return res.json();
}

export async function exportShopify(batchId: string, params: { store_domain: string; access_token: string; publish_status: string; only_approved: boolean }) {
  const res = await fetch(`${BASE_URL}/batches/${batchId}/export/shopify`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Export failed' }));
    throw new Error(err.error || 'Shopify export failed');
  }
  return res.json();
}
