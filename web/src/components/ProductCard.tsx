import React from 'react';
import { Check, X, Eye, ShieldCheck, AlertCircle, Maximize2 } from 'lucide-react';
import { Product } from '../types';

interface ProductCardProps {
  product: Product;
  onInspect: (p: Product) => void;
  onUpdateStatus: (p: Product, status: 'APPROVED' | 'REJECTED') => void;
}

export const ProductCard: React.FC<ProductCardProps> = ({
  product,
  onInspect,
  onUpdateStatus,
}) => {
  const isApproved = product.publication_status === 'APPROVED';
  const isRejected = product.publication_status === 'REJECTED';

  // Compliance color
  const complianceBorder =
    product.image_compliance_status === 'PASS'
      ? 'border-emerald-500/20 hover:border-emerald-500/40'
      : product.image_compliance_status === 'WARN'
      ? 'border-amber-500/20 hover:border-amber-500/40'
      : 'border-rose-500/30 hover:border-rose-500/50';

  const scoreColor =
    product.quality_score >= 90
      ? 'text-emerald-400 border-emerald-500/30 bg-emerald-500/10'
      : product.quality_score >= 75
      ? 'text-amber-400 border-amber-500/30 bg-amber-500/10'
      : 'text-rose-400 border-rose-500/30 bg-rose-500/10';

  return (
    <div
      className={`group bg-slate-900 border ${complianceBorder} rounded-xl overflow-hidden flex flex-col justify-between transition duration-200 hover:shadow-xl hover:shadow-slate-950/50`}
    >
      <div>
        {/* Image & Overlay Indicators */}
        <div className="relative aspect-square bg-slate-950 overflow-hidden flex items-center justify-center border-b border-slate-800/80">
          {product.image_url ? (
            <img
              src={product.image_url}
              alt={product.raw_title}
              className="w-full h-full object-contain p-2 group-hover:scale-105 transition duration-300"
              loading="lazy"
              onError={(e) => {
                // Fallback for broken images
                (e.target as HTMLImageElement).src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="%23475569" stroke-width="1.5"><rect width="18" height="18" x="3" y="3" rx="2" ry="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-3.086-3.086a2 2 0 0 0-2.828 0L6 21"/></svg>';
              }}
            />
          ) : (
            <div className="text-slate-600 text-xs flex flex-col items-center gap-1">
              <Eye className="w-5 h-5 text-slate-600" />
              <span>No Image</span>
            </div>
          )}

          {/* Top Compliance Floating Badges */}
          <div className="absolute top-2 left-2 flex flex-col gap-1 z-10">
            <span
              className={`text-[10px] font-bold px-2 py-0.5 rounded-full backdrop-blur-md border shadow-sm flex items-center gap-1 ${
                product.image_compliance_status === 'PASS'
                  ? 'bg-emerald-950/80 text-emerald-300 border-emerald-500/30'
                  : product.image_compliance_status === 'WARN'
                  ? 'bg-amber-950/80 text-amber-300 border-amber-500/30'
                  : 'bg-rose-950/80 text-rose-300 border-rose-500/30'
              }`}
            >
              {product.image_compliance_status === 'PASS' ? (
                <ShieldCheck className="w-3 h-3 text-emerald-400" />
              ) : (
                <AlertCircle className="w-3 h-3 text-amber-400" />
              )}
              <span>{(product.image_white_bg_score * 100).toFixed(1)}% White</span>
            </span>

            {product.image_zoom_ready && (
              <span className="text-[9px] font-semibold px-1.5 py-0.5 rounded bg-slate-900/80 text-slate-300 border border-slate-700/50 backdrop-blur w-fit flex items-center gap-0.5">
                <Maximize2 className="w-2.5 h-2.5 text-sky-400" />
                Zoom Ready
              </span>
            )}
          </div>

          {/* Top Right Quality Score Radial / Pill */}
          <div className="absolute top-2 right-2 z-10">
            <span
              className={`text-xs font-mono font-bold px-2 py-0.5 rounded-md border backdrop-blur-md shadow-sm ${scoreColor}`}
              title="Listing Quality Score (0-100)"
            >
              {product.quality_score}%
            </span>
          </div>

          {/* Quick inspect button hover overlay */}
          <button
            onClick={() => onInspect(product)}
            className="absolute inset-0 bg-slate-950/40 opacity-0 group-hover:opacity-100 backdrop-blur-[2px] transition flex items-center justify-center gap-1.5 text-xs font-semibold text-white z-20"
          >
            <Eye className="w-4 h-4 text-sky-400" />
            <span>Audit & Edit Listing</span>
          </button>
        </div>

        {/* Content Body */}
        <div className="p-3.5 space-y-2.5">
          {/* Header Row: SKU & Brand */}
          <div className="flex items-center justify-between text-xs">
            <span className="font-mono text-[11px] text-sky-400 font-semibold bg-sky-950/50 px-1.5 py-0.5 rounded border border-sky-800/40">
              {product.raw_sku}
            </span>
            <span className="text-slate-400 font-medium truncate max-w-[120px]">
              {product.brand || 'Unbranded'}
            </span>
          </div>

          {/* Optimized Title */}
          <h4
            className="text-xs font-semibold text-slate-200 line-clamp-2 leading-snug group-hover:text-white transition cursor-pointer"
            onClick={() => onInspect(product)}
            title={product.optimized_title || product.raw_title}
          >
            {product.optimized_title || product.raw_title}
          </h4>

          {/* Extracted Attributes Chips */}
          <div className="flex items-center flex-wrap gap-1 pt-1">
            {product.extracted_attributes.color_hex && (
              <span
                className="w-3.5 h-3.5 rounded-full border border-slate-700 shadow-inner flex-shrink-0"
                style={{ backgroundColor: product.extracted_attributes.color_hex }}
                title={`Color: ${product.extracted_attributes.color_family || product.extracted_attributes.color_hex}`}
              />
            )}
            {product.extracted_attributes.material && (
              <span className="text-[10px] font-medium px-1.5 py-0.5 rounded bg-slate-800/80 text-slate-300 border border-slate-700/60 truncate max-w-[120px]">
                {product.extracted_attributes.material}
              </span>
            )}
            {product.extracted_attributes.fit && (
              <span className="text-[10px] font-medium px-1.5 py-0.5 rounded bg-slate-800/80 text-slate-300 border border-slate-700/60 truncate max-w-[110px]">
                {product.extracted_attributes.fit}
              </span>
            )}
            {product.extracted_attributes.collar && (
              <span className="text-[10px] font-medium px-1.5 py-0.5 rounded bg-slate-800/80 text-slate-300 border border-slate-700/60 truncate max-w-[110px]">
                {product.extracted_attributes.collar}
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Footer Actions: Status & Buttons */}
      <div className="px-3.5 py-2.5 bg-slate-950/60 border-t border-slate-800 flex items-center justify-between text-xs">
        <div className="font-semibold text-slate-200">
          ${product.raw_price.toFixed(2)}{' '}
          <span className="text-[10px] text-slate-500">{product.currency}</span>
        </div>

        <div className="flex items-center gap-1">
          <button
            onClick={() => onUpdateStatus(product, 'APPROVED')}
            className={`p-1.5 rounded-lg border transition ${
              isApproved
                ? 'bg-emerald-500 text-slate-950 border-emerald-400 font-bold'
                : 'bg-slate-800 hover:bg-emerald-500/20 text-slate-300 hover:text-emerald-400 border-slate-700'
            }`}
            title="Approve Listing"
          >
            <Check className="w-3.5 h-3.5" />
          </button>
          <button
            onClick={() => onUpdateStatus(product, 'REJECTED')}
            className={`p-1.5 rounded-lg border transition ${
              isRejected
                ? 'bg-rose-500 text-white border-rose-400 font-bold'
                : 'bg-slate-800 hover:bg-rose-500/20 text-slate-300 hover:text-rose-400 border-slate-700'
            }`}
            title="Reject Listing"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>
    </div>
  );
};
