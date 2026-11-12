import React, { useState, useRef } from 'react';
import { X, Upload, FileText, AlertCircle, Loader2 } from 'lucide-react';
import { uploadFeed } from '../lib/api';

interface UploadModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (batchId: string) => void;
}

export const UploadModal: React.FC<UploadModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  if (!isOpen) return null;

  const [file, setFile] = useState<File | null>(null);
  const [batchName, setBatchName] = useState('');
  const [isUploading, setIsUploading] = useState(false);
  const [progress, setProgress] = useState<{ current: number; total: number; percent: number; currentSKU: string } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileDrop = (e: React.DragEvent) => {
    e.preventDefault();
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const selected = e.dataTransfer.files[0];
      setFile(selected);
      if (!batchName) {
        setBatchName(selected.name.replace(/\.[^/.]+$/, ''));
      }
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const selected = e.target.files[0];
      setFile(selected);
      if (!batchName) {
        setBatchName(selected.name.replace(/\.[^/.]+$/, ''));
      }
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) return;

    setIsUploading(true);
    setError(null);

    try {
      const res = await uploadFeed(file, batchName || file.name);
      const batchId = res.batch_id;

      // Subscribe to Server-Sent Events for real-time progress
      const eventSource = new EventSource(`/api/batches/${batchId}/events`);

      eventSource.addEventListener('progress', (e) => {
        try {
          const data = JSON.parse(e.data);
          setProgress({
            current: data.current,
            total: data.total,
            percent: data.percent,
            currentSKU: data.current_sku,
          });
        } catch {}
      });

      eventSource.addEventListener('complete', () => {
        eventSource.close();
        setIsUploading(false);
        onSuccess(batchId);
        onClose();
      });

      eventSource.onerror = () => {
        eventSource.close();
        // If SSE closes or finishes, trigger success after brief pause
        setTimeout(() => {
          setIsUploading(false);
          onSuccess(batchId);
          onClose();
        }, 1500);
      };
    } catch (err: any) {
      setError(err.message || 'Failed to upload feed');
      setIsUploading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-950/50">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-sky-500/10 text-sky-400 flex items-center justify-center border border-sky-500/20">
              <Upload className="w-4 h-4" />
            </div>
            <div>
              <h3 className="font-bold text-slate-100 text-sm">Import Supplier Feed</h3>
              <p className="text-xs text-slate-400">CSV, TSV, or JSON product catalog</p>
            </div>
          </div>
          <button
            onClick={onClose}
            disabled={isUploading}
            className="p-1 rounded-lg hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition disabled:opacity-50"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs flex items-center gap-2">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Drag & Drop File Zone */}
          <div
            onDragOver={(e) => e.preventDefault()}
            onDrop={handleFileDrop}
            onClick={() => fileInputRef.current?.click()}
            className="border-2 border-dashed border-slate-800 hover:border-sky-500/50 rounded-xl p-8 flex flex-col items-center justify-center gap-3 bg-slate-950/40 cursor-pointer transition group"
          >
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileSelect}
              accept=".csv,.tsv,.json"
              className="hidden"
            />
            <div className="w-12 h-12 rounded-xl bg-slate-800 group-hover:bg-sky-500/10 text-slate-400 group-hover:text-sky-400 flex items-center justify-center transition">
              <FileText className="w-6 h-6" />
            </div>
            <div className="text-center">
              {file ? (
                <div className="text-xs font-semibold text-sky-400">{file.name}</div>
              ) : (
                <>
                  <div className="text-xs font-semibold text-slate-200">
                    Click to upload or drag and drop
                  </div>
                  <div className="text-[11px] text-slate-500 mt-1">
                    CSV, TSV, or JSON supplier files
                  </div>
                </>
              )}
            </div>
          </div>

          {/* Batch Name Input */}
          <div>
            <label className="text-xs font-medium text-slate-400 block mb-1.5">
              Batch Name
            </label>
            <input
              type="text"
              value={batchName}
              onChange={(e) => setBatchName(e.target.value)}
              placeholder="e.g. Summer 2026 Apparel Supplier Feed"
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 focus:border-sky-500 outline-none"
            />
          </div>

          {/* Live Progress Bar during Enrichment */}
          {isUploading && progress && (
            <div className="p-3.5 rounded-xl bg-slate-950 border border-slate-800 space-y-2">
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-400 flex items-center gap-1.5">
                  <Loader2 className="w-3.5 h-3.5 animate-spin text-sky-400" />
                  Auditing & Enriching: <span className="font-mono text-slate-200">{progress.currentSKU}</span>
                </span>
                <span className="font-mono text-sky-400 font-bold">{progress.percent}%</span>
              </div>
              <div className="w-full h-2 bg-slate-800 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-sky-500 to-indigo-500 transition-all duration-300"
                  style={{ width: `${progress.percent}%` }}
                />
              </div>
              <div className="text-[10px] text-slate-500 text-right">
                {progress.current} of {progress.total} SKUs
              </div>
            </div>
          )}

          {/* Footer Actions */}
          <div className="pt-2 flex items-center justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              disabled={isUploading}
              className="px-4 py-2 text-xs font-medium rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 transition disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!file || isUploading}
              className="flex items-center gap-1.5 px-4 py-2 text-xs font-semibold rounded-lg bg-sky-500 hover:bg-sky-400 text-slate-950 transition shadow-lg shadow-sky-500/20 disabled:opacity-50"
            >
              {isUploading ? (
                <>
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  <span>Processing Batch...</span>
                </>
              ) : (
                <>
                  <Upload className="w-3.5 h-3.5" />
                  <span>Ingest & Audit Catalog</span>
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
