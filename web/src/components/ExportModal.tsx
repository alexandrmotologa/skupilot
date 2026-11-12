import React, { useState } from 'react';
import { X, Share2, Download, ShoppingBag, Globe, FileSpreadsheet, Check, Loader2, AlertCircle } from 'lucide-react';
import { CatalogBatch } from '../types';
import { exportShopify } from '../lib/api';

interface ExportModalProps {
  isOpen: boolean;
  onClose: () => void;
  batch: CatalogBatch | null;
}

export const ExportModal: React.FC<ExportModalProps> = ({
  isOpen,
  onClose,
  batch,
}) => {
  if (!isOpen || !batch) return null;

  const [activeTab, setActiveTab] = useState<'shopify' | 'amazon' | 'google' | 'csv'>('shopify');
  const [storeDomain, setStoreDomain] = useState('');
  const [accessToken, setAccessToken] = useState('');
  const [publishStatus, setPublishStatus] = useState('DRAFT');
  const [onlyApproved, setOnlyApproved] = useState(true);
  const [isExporting, setIsExporting] = useState(false);
  const [result, setResult] = useState<{ status: string; count: number; message: string } | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleShopifySync = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsExporting(true);
    setError(null);
    setResult(null);

    try {
      const res = await exportShopify(batch.id, {
        store_domain: storeDomain,
        access_token: accessToken,
        publish_status: publishStatus,
        only_approved: onlyApproved,
      });

      setResult({
        status: res.status,
        count: res.exported_count,
        message: res.details || `Successfully exported ${res.exported_count} products to Shopify`,
      });
    } catch (err: any) {
      setError(err.message || 'Shopify export failed');
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-950/50">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-sky-500/10 text-sky-400 flex items-center justify-center border border-sky-500/20">
              <Share2 className="w-4 h-4" />
            </div>
            <div>
              <h3 className="font-bold text-slate-100 text-sm">Channel Exporter</h3>
              <p className="text-xs text-slate-400">Sync enriched SKUs to marketplaces</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Channel Selector Tabs */}
        <div className="px-6 pt-4 border-b border-slate-800 flex items-center gap-4 text-xs font-semibold">
          <button
            onClick={() => { setActiveTab('shopify'); setResult(null); }}
            className={`pb-3 flex items-center gap-1.5 border-b-2 transition ${
              activeTab === 'shopify' ? 'border-sky-500 text-sky-400' : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            <ShoppingBag className="w-4 h-4" />
            <span>Shopify GraphQL</span>
          </button>
          <button
            onClick={() => { setActiveTab('amazon'); setResult(null); }}
            className={`pb-3 flex items-center gap-1.5 border-b-2 transition ${
              activeTab === 'amazon' ? 'border-sky-500 text-sky-400' : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            <FileSpreadsheet className="w-4 h-4" />
            <span>Amazon TSV Flat File</span>
          </button>
          <button
            onClick={() => { setActiveTab('google'); setResult(null); }}
            className={`pb-3 flex items-center gap-1.5 border-b-2 transition ${
              activeTab === 'google' ? 'border-sky-500 text-sky-400' : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            <Globe className="w-4 h-4" />
            <span>Google Shopping XML</span>
          </button>
          <button
            onClick={() => { setActiveTab('csv'); setResult(null); }}
            className={`pb-3 flex items-center gap-1.5 border-b-2 transition ${
              activeTab === 'csv' ? 'border-sky-500 text-sky-400' : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            <Download className="w-4 h-4" />
            <span>Enriched PIM CSV</span>
          </button>
        </div>

        {/* Tab Content */}
        <div className="p-6">
          {error && (
            <div className="p-3 mb-4 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs flex items-center gap-2">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {result && (
            <div className="p-3.5 mb-4 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs space-y-1">
              <div className="font-bold flex items-center gap-1.5">
                <Check className="w-4 h-4" />
                <span>Export Successful ({result.status})</span>
              </div>
              <p className="text-slate-300 text-[11px]">{result.message}</p>
            </div>
          )}

          {/* Shopify Tab */}
          {activeTab === 'shopify' && (
            <form onSubmit={handleShopifySync} className="space-y-4 text-xs">
              <div>
                <label className="text-slate-400 block mb-1 font-medium">Store Domain</label>
                <input
                  type="text"
                  placeholder="e.g. brand-name.myshopify.com"
                  value={storeDomain}
                  onChange={(e) => setStoreDomain(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 focus:border-sky-500 outline-none"
                />
              </div>

              <div>
                <label className="text-slate-400 block mb-1 font-medium">
                  Admin Access Token{' '}
                  <span className="text-slate-500 font-normal">(leave blank for Dry Run)</span>
                </label>
                <input
                  type="password"
                  placeholder="shpat_xxxxxxxxxxxxxxxxxxxx"
                  value={accessToken}
                  onChange={(e) => setAccessToken(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 font-mono focus:border-sky-500 outline-none"
                />
              </div>

              <div className="grid grid-cols-2 gap-3 pt-1">
                <div>
                  <label className="text-slate-400 block mb-1 font-medium">Product Status</label>
                  <select
                    value={publishStatus}
                    onChange={(e) => setPublishStatus(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-slate-200 outline-none"
                  >
                    <option value="DRAFT">DRAFT (Safe)</option>
                    <option value="ACTIVE">ACTIVE (Live)</option>
                  </select>
                </div>

                <div className="flex items-center pt-5">
                  <label className="flex items-center gap-2 text-slate-300 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={onlyApproved}
                      onChange={(e) => setOnlyApproved(e.target.checked)}
                      className="rounded bg-slate-950 border-slate-800 text-sky-500 focus:ring-0"
                    />
                    <span>Only Approved SKUs</span>
                  </label>
                </div>
              </div>

              <div className="pt-3 flex items-center justify-end gap-3">
                <button
                  type="button"
                  onClick={onClose}
                  className="px-4 py-2 font-medium rounded-lg bg-slate-800 text-slate-200"
                >
                  Close
                </button>
                <button
                  type="submit"
                  disabled={isExporting}
                  className="flex items-center gap-1.5 px-4 py-2 font-semibold rounded-lg bg-sky-500 hover:bg-sky-400 text-slate-950 transition shadow-lg shadow-sky-500/20 disabled:opacity-50"
                >
                  {isExporting ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <ShoppingBag className="w-3.5 h-3.5" />}
                  <span>{accessToken ? 'Publish to Shopify' : 'Validate Dry Run'}</span>
                </button>
              </div>
            </form>
          )}

          {/* Amazon Tab */}
          {activeTab === 'amazon' && (
            <div className="space-y-4 text-xs">
              <p className="text-slate-400">
                Generates a tab-separated inventory loader file conforming to Amazon Seller Central style guides. Includes formulaic A9 titles, 5 bullets, and image URLs.
              </p>
              <div className="p-3 rounded-xl bg-slate-950 border border-slate-800 space-y-1">
                <div className="text-slate-300 font-semibold">Amazon Inventory Loader TSV</div>
                <div className="text-[11px] text-slate-500">
                  Ready to upload directly into Seller Central &gt; Add Products via Upload
                </div>
              </div>
              <div className="pt-2 flex justify-end">
                <a
                  href={`/api/batches/${batch.id}/export/amazon`}
                  download
                  className="flex items-center gap-1.5 px-4 py-2 font-semibold rounded-lg bg-sky-500 hover:bg-sky-400 text-slate-950 transition shadow-lg shadow-sky-500/20"
                >
                  <Download className="w-3.5 h-3.5" />
                  <span>Download Amazon TSV</span>
                </a>
              </div>
            </div>
          )}

          {/* Google Shopping Tab */}
          {activeTab === 'google' && (
            <div className="space-y-4 text-xs">
              <p className="text-slate-400">
                Generates a standard RSS 2.0 Google Merchant Center product feed with official numeric Google Taxonomy IDs, high-resolution image links, and standardized attributes.
              </p>
              <div className="p-3 rounded-xl bg-slate-950 border border-slate-800 space-y-1">
                <div className="text-slate-300 font-semibold">Google Merchant Center XML</div>
                <div className="text-[11px] text-slate-500">
                  Compatible with Google Merchant Center Primary Feed schedules
                </div>
              </div>
              <div className="pt-2 flex justify-end">
                <a
                  href={`/api/batches/${batch.id}/export/google`}
                  download
                  className="flex items-center gap-1.5 px-4 py-2 font-semibold rounded-lg bg-sky-500 hover:bg-sky-400 text-slate-950 transition shadow-lg shadow-sky-500/20"
                >
                  <Download className="w-3.5 h-3.5" />
                  <span>Download Google XML</span>
                </a>
              </div>
            </div>
          )}

          {/* Universal Enriched CSV Tab */}
          {activeTab === 'csv' && (
            <div className="space-y-4 text-xs">
              <p className="text-slate-400">
                Exports all enriched attributes, visual whiteness scores, A9 titles, 5 bullets, and Google Taxonomy IDs into a standardized Master PIM CSV.
              </p>
              <div className="p-3 rounded-xl bg-slate-950 border border-slate-800 space-y-1">
                <div className="text-slate-300 font-semibold">Master PIM & ERP Enriched Feed</div>
                <div className="text-[11px] text-slate-500">
                  Compatible with Akeneo, Pimcore, Magento, WooCommerce, and custom databases
                </div>
              </div>
              <div className="pt-2 flex justify-end">
                <a
                  href={`/api/batches/${batch.id}/export/csv`}
                  download
                  className="flex items-center gap-1.5 px-4 py-2 font-semibold rounded-lg bg-sky-500 hover:bg-sky-400 text-slate-950 transition shadow-lg shadow-sky-500/20"
                >
                  <Download className="w-3.5 h-3.5" />
                  <span>Download Master CSV</span>
                </a>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
