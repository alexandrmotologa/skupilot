import React, { useState, useEffect, useCallback } from 'react';
import { Header } from './components/Header';
import { FilterBar } from './components/FilterBar';
import { ProductCard } from './components/ProductCard';
import { ProductDetailModal } from './components/ProductDetailModal';
import { UploadModal } from './components/UploadModal';
import { ExportModal } from './components/ExportModal';
import { AnalyticsDashboard } from './components/AnalyticsDashboard';
import { CatalogBatch, Product, BatchAnalytics } from './types';
import {
  fetchBatches,
  fetchProducts,
  updateProduct,
  approveAllBatchProducts,
  fetchBatchAnalytics,
  bulkUpdateStatus,
  autoFixProductImage,
} from './lib/api';
import { PackageSearch, Sparkles, CheckSquare, Square, Check, X, ShieldCheck, XCircle } from 'lucide-react';

export const App: React.FC = () => {
  const [batches, setBatches] = useState<CatalogBatch[]>([]);
  const [currentBatch, setCurrentBatch] = useState<CatalogBatch | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [analytics, setAnalytics] = useState<BatchAnalytics | null>(null);
  const [isAnalyticsOpen, setIsAnalyticsOpen] = useState(true);

  // Filters
  const [search, setSearch] = useState('');
  const [complianceFilter, setComplianceFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  // Bulk Selection
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  // Modals
  const [inspectedProduct, setInspectedProduct] = useState<Product | null>(null);
  const [isUploadOpen, setIsUploadOpen] = useState(false);
  const [isExportOpen, setIsExportOpen] = useState(false);

  // Load Batches
  const loadBatches = useCallback(async (selectBatchId?: string) => {
    try {
      const data = await fetchBatches();
      setBatches(data);
      if (data.length > 0) {
        if (selectBatchId) {
          const found = data.find((b) => b.id === selectBatchId);
          if (found) setCurrentBatch(found);
        } else if (!currentBatch) {
          setCurrentBatch(data[0]);
        }
      }
    } catch (err) {
      console.error('Failed loading batches', err);
    }
  }, [currentBatch]);

  // Load Analytics
  const loadAnalytics = useCallback(async () => {
    if (!currentBatch) return;
    try {
      const data = await fetchBatchAnalytics(currentBatch.id);
      setAnalytics(data);
    } catch (err) {
      console.error('Failed loading analytics', err);
    }
  }, [currentBatch]);

  // Load Products for current batch
  const loadProducts = useCallback(async () => {
    if (!currentBatch) return;
    setIsLoading(true);
    try {
      const res = await fetchProducts(currentBatch.id, {
        search,
        compliance: complianceFilter,
        status: statusFilter,
        limit: 100,
      });
      setProducts(res.items);
      setTotalCount(res.total);
    } catch (err) {
      console.error('Failed loading products', err);
    } finally {
      setIsLoading(false);
    }
  }, [currentBatch, search, complianceFilter, statusFilter]);

  useEffect(() => {
    loadBatches();
  }, [loadBatches]);

  useEffect(() => {
    loadProducts();
    loadAnalytics();
  }, [loadProducts, loadAnalytics]);

  // Selection handlers
  const handleToggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const handleSelectAll = () => {
    if (selectedIds.size === products.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(products.map((p) => p.id)));
    }
  };

  const handleSelectCompliant = () => {
    const compliant = products.filter((p) => p.image_compliance_status === 'PASS').map((p) => p.id);
    setSelectedIds(new Set(compliant));
  };

  const handleBulkStatusChange = async (status: 'APPROVED' | 'REJECTED') => {
    if (!currentBatch || selectedIds.size === 0) return;
    try {
      await bulkUpdateStatus(currentBatch.id, Array.from(selectedIds), status);
      await loadProducts();
      await loadAnalytics();
      setSelectedIds(new Set());
    } catch (err) {
      console.error('Failed bulk status update', err);
    }
  };

  const handleUpdateStatus = async (product: Product, status: 'APPROVED' | 'REJECTED') => {
    try {
      const updated = await updateProduct(product.id, { publication_status: status });
      setProducts((prev) => prev.map((p) => (p.id === updated.id ? updated : p)));
      await loadAnalytics();
    } catch (err) {
      console.error('Failed to update status', err);
    }
  };

  const handleAutoFixImage = async (product: Product) => {
    try {
      const res = await autoFixProductImage(product.id);
      setProducts((prev) => prev.map((p) => (p.id === res.product.id ? res.product : p)));
      await loadAnalytics();
    } catch (err) {
      console.error('Failed auto fixing image', err);
    }
  };

  const handleSaveProduct = async (updated: Product) => {
    try {
      const res = await updateProduct(updated.id, updated);
      setProducts((prev) => prev.map((p) => (p.id === res.id ? res : p)));
      await loadAnalytics();
    } catch (err) {
      console.error('Failed to save product updates', err);
    }
  };

  const handleApproveAll = async () => {
    if (!currentBatch) return;
    try {
      await approveAllBatchProducts(currentBatch.id);
      await loadProducts();
      await loadBatches(currentBatch.id);
      await loadAnalytics();
    } catch (err) {
      console.error('Failed to approve batch', err);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans pb-16">
      <Header
        batches={batches}
        currentBatch={currentBatch}
        onSelectBatch={(b) => {
          setCurrentBatch(b);
          setSelectedIds(new Set());
        }}
        onOpenUpload={() => setIsUploadOpen(true)}
        onOpenExport={() => setIsExportOpen(true)}
        onApproveAll={handleApproveAll}
        stats={null}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto px-6 py-6 space-y-6">
        {/* Catalog Analytics Dashboard */}
        <AnalyticsDashboard
          analytics={analytics}
          isOpen={isAnalyticsOpen}
          onToggle={() => setIsAnalyticsOpen(!isAnalyticsOpen)}
          onFilterIssue={(issue) => {
            if (issue === 'non_white') setComplianceFilter('FAIL');
            else if (issue === 'low_res') setComplianceFilter('WARN');
          }}
        />

        {/* Filter & Search Toolbar */}
        <FilterBar
          search={search}
          onSearchChange={setSearch}
          complianceFilter={complianceFilter}
          onComplianceFilterChange={setComplianceFilter}
          statusFilter={statusFilter}
          onStatusFilterChange={setStatusFilter}
          totalFiltered={totalCount}
        />

        {/* Bulk Action Header when items selected */}
        {products.length > 0 && (
          <div className="flex items-center justify-between text-xs px-2 py-1 text-slate-400">
            <div className="flex items-center gap-3">
              <button
                onClick={handleSelectAll}
                className="flex items-center gap-1.5 hover:text-slate-200 transition font-medium"
              >
                {selectedIds.size === products.length ? (
                  <CheckSquare className="w-4 h-4 text-sky-400" />
                ) : (
                  <Square className="w-4 h-4 text-slate-500" />
                )}
                <span>Select All ({products.length})</span>
              </button>

              <button
                onClick={handleSelectCompliant}
                className="flex items-center gap-1.5 hover:text-slate-200 transition font-medium"
              >
                <ShieldCheck className="w-3.5 h-3.5 text-emerald-400" />
                <span>Select Compliant</span>
              </button>
            </div>

            {selectedIds.size > 0 && (
              <div className="flex items-center gap-2 animate-in fade-in">
                <span className="font-semibold text-sky-400 font-mono">
                  {selectedIds.size} selected
                </span>
                <button
                  onClick={() => handleBulkStatusChange('APPROVED')}
                  className="flex items-center gap-1 px-2.5 py-1 rounded bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20 border border-emerald-500/20 font-semibold"
                >
                  <Check className="w-3 h-3" /> Approve Selected
                </button>
                <button
                  onClick={() => handleBulkStatusChange('REJECTED')}
                  className="flex items-center gap-1 px-2.5 py-1 rounded bg-rose-500/10 text-rose-400 hover:bg-rose-500/20 border border-rose-500/20 font-semibold"
                >
                  <X className="w-3 h-3" /> Reject Selected
                </button>
                <button
                  onClick={() => setSelectedIds(new Set())}
                  className="p-1 text-slate-500 hover:text-slate-300"
                  title="Clear Selection"
                >
                  <XCircle className="w-4 h-4" />
                </button>
              </div>
            )}
          </div>
        )}

        {/* Product Cards Grid */}
        {isLoading ? (
          <div className="py-24 flex flex-col items-center justify-center gap-3 text-slate-400">
            <div className="w-8 h-8 border-2 border-sky-500 border-t-transparent rounded-full animate-spin" />
            <span className="text-xs">Loading enriched catalog...</span>
          </div>
        ) : products.length > 0 ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {products.map((p) => (
              <ProductCard
                key={p.id}
                product={p}
                isSelected={selectedIds.has(p.id)}
                onToggleSelect={handleToggleSelect}
                onInspect={(prod) => setInspectedProduct(prod)}
                onUpdateStatus={handleUpdateStatus}
                onAutoFixImage={handleAutoFixImage}
              />
            ))}
          </div>
        ) : (
          <div className="py-24 border border-slate-800/80 rounded-2xl bg-slate-900/30 flex flex-col items-center justify-center gap-3 text-center p-6">
            <div className="w-12 h-12 rounded-2xl bg-slate-800/80 text-slate-400 flex items-center justify-center">
              <PackageSearch className="w-6 h-6 text-slate-500" />
            </div>
            <div>
              <h4 className="font-semibold text-slate-200 text-sm">No Products Found</h4>
              <p className="text-xs text-slate-400 mt-1 max-w-sm">
                {batches.length === 0
                  ? 'Get started by uploading a raw supplier CSV or JSON catalog feed.'
                  : 'No products match the selected filters or search terms.'}
              </p>
            </div>
            {batches.length === 0 && (
              <button
                onClick={() => setIsUploadOpen(true)}
                className="mt-2 flex items-center gap-1.5 px-3.5 py-2 text-xs font-semibold rounded-lg bg-sky-500 hover:bg-sky-400 text-slate-950 transition"
              >
                <Sparkles className="w-3.5 h-3.5" />
                <span>Import First Feed</span>
              </button>
            )}
          </div>
        )}
      </main>

      {/* Modals */}
      <ProductDetailModal
        product={inspectedProduct}
        onClose={() => setInspectedProduct(null)}
        onSave={handleSaveProduct}
      />

      <UploadModal
        isOpen={isUploadOpen}
        onClose={() => setIsUploadOpen(false)}
        onSuccess={(batchId) => {
          loadBatches(batchId);
        }}
      />

      <ExportModal
        isOpen={isExportOpen}
        onClose={() => setIsExportOpen(false)}
        batch={currentBatch}
      />
    </div>
  );
};

export default App;
