import React from 'react';
import { Sparkles, Upload, Share2, CheckCircle2, Layers } from 'lucide-react';
import { CatalogBatch } from '../types';

interface HeaderProps {
  batches: CatalogBatch[];
  currentBatch: CatalogBatch | null;
  onSelectBatch: (batch: CatalogBatch) => void;
  onOpenUpload: () => void;
  onOpenExport: () => void;
  onApproveAll: () => void;
  stats: { total_products: number; compliant_products: number } | null;
}

export const Header: React.FC<HeaderProps> = ({
  batches,
  currentBatch,
  onSelectBatch,
  onOpenUpload,
  onOpenExport,
  onApproveAll,
}) => {
  return (
    <header className="border-b border-slate-800 bg-slate-900/90 backdrop-blur sticky top-0 z-40 px-6 py-4">
      <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
        {/* Brand & Batch Selector */}
        <div className="flex items-center gap-4 w-full md:w-auto">
          <div className="flex items-center gap-2.5">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-sky-500 to-indigo-600 flex items-center justify-center shadow-lg shadow-sky-500/20">
              <Sparkles className="w-5 h-5 text-white" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-bold text-lg text-white tracking-tight">skupilot</span>
                <span className="text-[10px] uppercase font-semibold px-1.5 py-0.5 rounded bg-sky-500/10 text-sky-400 border border-sky-500/20">
                  Multimodal VLM
                </span>
              </div>
              <p className="text-xs text-slate-400">Autonomous Catalog Merchandising</p>
            </div>
          </div>

          <div className="h-6 w-px bg-slate-800 hidden sm:block" />

          {/* Batch Selector Dropdown */}
          <div className="relative flex-1 sm:flex-initial">
            <div className="flex items-center gap-2 bg-slate-950 border border-slate-800 rounded-lg px-3 py-1.5 text-xs text-slate-200">
              <Layers className="w-3.5 h-3.5 text-slate-400" />
              <select
                className="bg-transparent border-none outline-none text-slate-200 cursor-pointer pr-4 font-medium"
                value={currentBatch?.id || ''}
                onChange={(e) => {
                  const b = batches.find((item) => item.id === e.target.value);
                  if (b) onSelectBatch(b);
                }}
              >
                {batches.length === 0 && <option value="">No batches loaded</option>}
                {batches.map((b) => (
                  <option key={b.id} value={b.id} className="bg-slate-900 text-slate-200">
                    {b.name} ({b.total_skus} SKUs)
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Global Batch Metrics & Action Buttons */}
        <div className="flex items-center gap-3 w-full md:w-auto justify-end">
          {currentBatch && (
            <div className="hidden lg:flex items-center gap-4 px-3 py-1.5 rounded-lg bg-slate-950/60 border border-slate-800 text-xs">
              <div>
                <span className="text-slate-500">Total: </span>
                <span className="font-semibold text-slate-200">{currentBatch.total_skus}</span>
              </div>
              <div className="w-px h-3.5 bg-slate-800" />
              <div>
                <span className="text-slate-500">Compliant: </span>
                <span className="font-semibold text-emerald-400">{currentBatch.compliant_count}</span>
              </div>
              <div className="w-px h-3.5 bg-slate-800" />
              <div>
                <span className="text-slate-500">Status: </span>
                <span className="font-medium text-sky-400">{currentBatch.status}</span>
              </div>
            </div>
          )}

          <button
            onClick={onApproveAll}
            disabled={!currentBatch}
            className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded-lg bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20 border border-emerald-500/20 transition disabled:opacity-50"
            title="Approve all compliant products for marketplace sync"
          >
            <CheckCircle2 className="w-3.5 h-3.5" />
            <span>Approve Compliant</span>
          </button>

          <button
            onClick={onOpenExport}
            disabled={!currentBatch}
            className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition disabled:opacity-50"
          >
            <Share2 className="w-3.5 h-3.5 text-slate-400" />
            <span>Sync Channels</span>
          </button>

          <button
            onClick={onOpenUpload}
            className="flex items-center gap-1.5 px-3.5 py-2 text-xs font-semibold rounded-lg bg-sky-500 hover:bg-sky-400 text-slate-950 shadow-md shadow-sky-500/20 transition"
          >
            <Upload className="w-3.5 h-3.5" />
            <span>Import Feed</span>
          </button>
        </div>
      </div>
    </header>
  );
};
