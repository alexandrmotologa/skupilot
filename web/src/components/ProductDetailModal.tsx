import React, { useState, useEffect } from 'react';
import { X, ShieldCheck, Save, RefreshCw, Sparkles, Layers, Maximize2, Wand2, ShoppingBag, Globe, Zap } from 'lucide-react';
import { Product, StyleVariant } from '../types';
import { autoFixProductImage, fetchCopyStyles } from '../lib/api';

interface ProductDetailModalProps {
  product: Product | null;
  onClose: () => void;
  onSave: (updated: Product) => Promise<void>;
}

export const ProductDetailModal: React.FC<ProductDetailModalProps> = ({
  product,
  onClose,
  onSave,
}) => {
  if (!product) return null;

  const [currentProduct, setCurrentProduct] = useState<Product>(product);
  const [title, setTitle] = useState(product.optimized_title || product.raw_title);
  const [bullets, setBullets] = useState<string[]>(
    product.bullet_points.length === 5
      ? product.bullet_points
      : [...product.bullet_points, '', '', '', '', ''].slice(0, 5)
  );
  const [attributes, setAttributes] = useState(product.extracted_attributes);
  const [status, setStatus] = useState(product.publication_status);
  const [isSaving, setIsSaving] = useState(false);
  const [isAutoFixing, setIsAutoFixing] = useState(false);
  const [activeTab, setActiveTab] = useState<'copy' | 'html'>('copy');
  const [activeCopyStyle, setActiveCopyStyle] = useState<'AMAZON_A9' | 'SHOPIFY_EDITORIAL' | 'GOOGLE_SHOPPING'>('AMAZON_A9');
  const [styleVariants, setStyleVariants] = useState<Record<string, StyleVariant> | null>(null);

  useEffect(() => {
    setCurrentProduct(product);
    setTitle(product.optimized_title || product.raw_title);
    setBullets(product.bullet_points.length === 5 ? product.bullet_points : [...product.bullet_points, '', '', '', '', ''].slice(0, 5));
    setAttributes(product.extracted_attributes);
    setStatus(product.publication_status);

    // Fetch copy style variants
    fetchCopyStyles(product.id)
      .then((data) => setStyleVariants(data))
      .catch((err) => console.error('Failed loading styles', err));
  }, [product]);

  const titleLength = title.length;
  const isTitleOptimal = titleLength >= 150 && titleLength <= 200;
  const isTitleAcceptable = titleLength >= 80 && titleLength <= 200;

  const handleBulletChange = (index: number, val: string) => {
    const next = [...bullets];
    next[index] = val;
    setBullets(next);
  };

  const handleAttributeChange = (key: keyof typeof attributes, val: string) => {
    setAttributes((prev) => ({ ...prev, [key]: val }));
  };

  const handleApplyStyle = (styleKey: 'AMAZON_A9' | 'SHOPIFY_EDITORIAL' | 'GOOGLE_SHOPPING') => {
    setActiveCopyStyle(styleKey);
    if (styleVariants && styleVariants[styleKey]) {
      const v = styleVariants[styleKey];
      setTitle(v.title);
      setBullets(v.bullets);
    }
  };

  const handleAutoFix = async () => {
    setIsAutoFixing(true);
    try {
      const res = await autoFixProductImage(currentProduct.id);
      setCurrentProduct(res.product);
    } catch (err) {
      console.error('Failed auto-fixing image', err);
    } finally {
      setIsAutoFixing(false);
    }
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      const updated: Product = {
        ...currentProduct,
        optimized_title: title,
        bullet_points: bullets,
        extracted_attributes: attributes,
        publication_status: status,
      };
      await onSave(updated);
      onClose();
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-5xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
        {/* Modal Header */}
        <div className="px-4 sm:px-6 py-4 border-b border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-slate-950/50">
          <div className="flex items-center gap-2.5 sm:gap-3 min-w-0">
            <span className="font-mono text-xs px-2 py-0.5 rounded bg-sky-500/10 text-sky-400 border border-sky-500/20 font-bold shrink-0">
              {currentProduct.raw_sku}
            </span>
            <h3 className="font-bold text-slate-100 text-sm truncate max-w-[200px] sm:max-w-md">
              {currentProduct.raw_title}
            </h3>
          </div>

          <div className="flex items-center justify-between sm:justify-end gap-3 shrink-0">
            {/* Quick Status Toggle */}
            <div className="flex items-center gap-1 bg-slate-950 p-1 rounded-lg border border-slate-800 text-xs shrink-0">
              <button
                onClick={() => setStatus('PENDING')}
                className={`px-2 py-0.5 rounded font-medium transition whitespace-nowrap shrink-0 ${
                  status === 'PENDING' ? 'bg-sky-500/20 text-sky-400' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                Pending
              </button>
              <button
                onClick={() => setStatus('APPROVED')}
                className={`px-2 py-0.5 rounded font-medium transition whitespace-nowrap shrink-0 ${
                  status === 'APPROVED' ? 'bg-emerald-500/20 text-emerald-400' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                Approved
              </button>
              <button
                onClick={() => setStatus('REJECTED')}
                className={`px-2 py-0.5 rounded font-medium transition whitespace-nowrap shrink-0 ${
                  status === 'REJECTED' ? 'bg-rose-500/20 text-rose-400' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                Rejected
              </button>
            </div>

            <button
              onClick={onClose}
              className="p-1 rounded-lg hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition shrink-0"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Modal Scrollable Body */}
        <div className="flex-1 overflow-y-auto p-6 grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Left Column: Image Audit & Attributes (5 cols) */}
          <div className="lg:col-span-5 space-y-5">
            {/* Image Preview & Zoom Container */}
            <div className="relative aspect-square rounded-xl bg-slate-950 border border-slate-800 overflow-hidden flex items-center justify-center p-2 group">
              <img
                src={currentProduct.image_url}
                alt={currentProduct.raw_title}
                className="w-full h-full object-contain"
              />
              <div className="absolute bottom-2 right-2 px-2 py-1 rounded bg-slate-900/90 border border-slate-800 text-[10px] text-slate-400 backdrop-blur font-mono flex items-center gap-1">
                <Maximize2 className="w-3 h-3 text-sky-400" />
                {currentProduct.image_width}x{currentProduct.image_height}px
              </div>
            </div>

            {/* Image Compliance Scorecard with Auto-Fix Button */}
            <div className="p-4 rounded-xl bg-slate-950/60 border border-slate-800 space-y-3">
              <div className="flex items-center justify-between">
                <h4 className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
                  <ShieldCheck className="w-4 h-4 text-sky-400" />
                  Marketplace Image Audit
                </h4>
                <span
                  className={`text-xs font-bold px-2 py-0.5 rounded-full border ${
                    currentProduct.image_compliance_status === 'PASS'
                      ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                      : 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                  }`}
                >
                  {currentProduct.image_compliance_status}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-2 text-xs">
                <div className="bg-slate-900 p-2.5 rounded-lg border border-slate-800">
                  <div className="text-slate-400 text-[11px]">RGB 255 Whiteness</div>
                  <div className="font-mono font-bold text-slate-100 text-sm mt-0.5">
                    {(currentProduct.image_white_bg_score * 100).toFixed(1)}%
                  </div>
                  <div className="text-[10px] text-slate-500 mt-0.5">Req: &gt;= 98.5%</div>
                </div>

                <div className="bg-slate-900 p-2.5 rounded-lg border border-slate-800">
                  <div className="text-slate-400 text-[11px]">Amazon Zoom Ready</div>
                  <div className="font-mono font-bold text-slate-100 text-sm mt-0.5 flex items-center gap-1">
                    {currentProduct.image_zoom_ready ? (
                      <span className="text-emerald-400 font-bold">Pass</span>
                    ) : (
                      <span className="text-rose-400 font-bold">Fails &lt;1000px</span>
                    )}
                  </div>
                  <div className="text-[10px] text-slate-500 mt-0.5">Min: 1000x1000</div>
                </div>
              </div>

              {/* 1-Click Auto-Remediate Image Button */}
              {currentProduct.image_compliance_status !== 'PASS' && (
                <button
                  onClick={handleAutoFix}
                  disabled={isAutoFixing}
                  className="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg bg-sky-500/10 hover:bg-sky-500/20 border border-sky-500/30 text-sky-400 text-xs font-bold transition disabled:opacity-50"
                >
                  {isAutoFixing ? (
                    <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <Wand2 className="w-3.5 h-3.5" />
                  )}
                  <span>Auto-Fix Background & Pad (1000x1000)</span>
                </button>
              )}
            </div>

            {/* Extracted Technical Attributes */}
            <div className="p-4 rounded-xl bg-slate-950/60 border border-slate-800 space-y-3">
              <div className="flex items-center justify-between">
                <h4 className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
                  <Sparkles className="w-4 h-4 text-sky-400" />
                  Extracted Attributes (VLM)
                </h4>
              </div>

              <div className="space-y-2 text-xs">
                <div>
                  <label className="text-slate-400 block mb-1">Fabric & Material</label>
                  <input
                    type="text"
                    value={attributes.material || ''}
                    onChange={(e) => handleAttributeChange('material', e.target.value)}
                    className="w-full bg-slate-900 border border-slate-800 rounded px-2.5 py-1.5 text-slate-200 focus:border-sky-500 outline-none"
                  />
                </div>

                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="text-slate-400 block mb-1">Fit & Cut</label>
                    <input
                      type="text"
                      value={attributes.fit || ''}
                      onChange={(e) => handleAttributeChange('fit', e.target.value)}
                      className="w-full bg-slate-900 border border-slate-800 rounded px-2.5 py-1.5 text-slate-200 focus:border-sky-500 outline-none"
                    />
                  </div>
                  <div>
                    <label className="text-slate-400 block mb-1">Collar / Neckline</label>
                    <input
                      type="text"
                      value={attributes.collar || attributes.neckline || ''}
                      onChange={(e) => handleAttributeChange('collar', e.target.value)}
                      className="w-full bg-slate-900 border border-slate-800 rounded px-2.5 py-1.5 text-slate-200 focus:border-sky-500 outline-none"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="text-slate-400 block mb-1">Color Family</label>
                    <input
                      type="text"
                      value={attributes.color_family || ''}
                      onChange={(e) => handleAttributeChange('color_family', e.target.value)}
                      className="w-full bg-slate-900 border border-slate-800 rounded px-2.5 py-1.5 text-slate-200 focus:border-sky-500 outline-none"
                    />
                  </div>
                  <div>
                    <label className="text-slate-400 block mb-1">Color Hex</label>
                    <div className="flex items-center gap-1.5">
                      <input
                        type="color"
                        value={attributes.color_hex || '#ffffff'}
                        onChange={(e) => handleAttributeChange('color_hex', e.target.value)}
                        className="w-7 h-7 rounded border border-slate-700 bg-transparent cursor-pointer"
                      />
                      <input
                        type="text"
                        value={attributes.color_hex || ''}
                        onChange={(e) => handleAttributeChange('color_hex', e.target.value)}
                        className="flex-1 bg-slate-900 border border-slate-800 rounded px-2 py-1 text-slate-200 font-mono text-xs focus:border-sky-500 outline-none uppercase"
                      />
                    </div>
                  </div>
                </div>

                <div>
                  <label className="text-slate-400 block mb-1">Closure / Fastening</label>
                  <input
                    type="text"
                    value={attributes.closure || ''}
                    onChange={(e) => handleAttributeChange('closure', e.target.value)}
                    className="w-full bg-slate-900 border border-slate-800 rounded px-2.5 py-1.5 text-slate-200 focus:border-sky-500 outline-none"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Right Column: Listing Copy & Multi-Style Selector (7 cols) */}
          <div className="lg:col-span-7 space-y-5">
            {/* Style Variant Switcher Bar */}
            <div className="p-2 rounded-xl bg-slate-950 border border-slate-800 flex flex-wrap sm:flex-nowrap items-center justify-between gap-2 text-xs">
              <span className="text-slate-400 font-semibold px-2 uppercase tracking-wider text-[10px] whitespace-nowrap">
                Copywriting Style:
              </span>
              <div className="flex items-center gap-1 overflow-x-auto">
                <button
                  onClick={() => handleApplyStyle('AMAZON_A9')}
                  className={`px-2.5 py-1 rounded-lg flex items-center gap-1 font-semibold transition whitespace-nowrap shrink-0 ${
                    activeCopyStyle === 'AMAZON_A9'
                      ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <Zap className="w-3 h-3 text-amber-400 shrink-0" /> Amazon A9
                </button>
                <button
                  onClick={() => handleApplyStyle('SHOPIFY_EDITORIAL')}
                  className={`px-2.5 py-1 rounded-lg flex items-center gap-1 font-semibold transition whitespace-nowrap shrink-0 ${
                    activeCopyStyle === 'SHOPIFY_EDITORIAL'
                      ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <ShoppingBag className="w-3 h-3 text-emerald-400 shrink-0" /> Shopify D2C
                </button>
                <button
                  onClick={() => handleApplyStyle('GOOGLE_SHOPPING')}
                  className={`px-2.5 py-1 rounded-lg flex items-center gap-1 font-semibold transition whitespace-nowrap shrink-0 ${
                    activeCopyStyle === 'GOOGLE_SHOPPING'
                      ? 'bg-sky-500/20 text-sky-300 border border-sky-500/30'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <Globe className="w-3 h-3 text-sky-400 shrink-0" /> Google Specs
                </button>
              </div>
            </div>

            {/* Google Taxonomy Breadcrumb */}
            <div className="px-3 py-2 rounded-lg bg-slate-950 border border-slate-800 text-xs flex items-center gap-2">
              <Layers className="w-3.5 h-3.5 text-sky-400 flex-shrink-0" />
              <span className="text-slate-400">Google Category:</span>
              <span className="text-slate-200 font-medium truncate">
                {currentProduct.google_taxonomy_path || 'Apparel & Accessories'}
              </span>
              <span className="font-mono text-slate-500 text-[10px]">
                (ID {currentProduct.google_taxonomy_id})
              </span>
            </div>

            {/* Title Synthesizer */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="text-xs font-bold uppercase tracking-wider text-slate-300">
                  Product Title
                </label>
                <div className="flex items-center gap-1.5 font-mono text-xs">
                  <span
                    className={
                      isTitleOptimal
                        ? 'text-emerald-400 font-bold'
                        : isTitleAcceptable
                        ? 'text-amber-400'
                        : 'text-rose-400'
                    }
                  >
                    {titleLength}
                  </span>
                  <span className="text-slate-500">/ 200 chars (target: 150-180)</span>
                </div>
              </div>
              <textarea
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                rows={3}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-sm text-slate-100 focus:border-sky-500 focus:outline-none transition leading-relaxed"
              />
            </div>

            {/* 5 Feature Bullets */}
            <div className="space-y-3">
              <label className="text-xs font-bold uppercase tracking-wider text-slate-300 flex items-center justify-between">
                <span>Feature Bullets & Benefits</span>
                <span className="text-[11px] font-normal text-slate-400">
                  {bullets.length} Bullets Active
                </span>
              </label>

              <div className="space-y-2">
                {bullets.map((b, idx) => (
                  <div key={idx} className="flex items-start gap-2">
                    <span className="w-5 h-5 rounded-full bg-slate-800 text-slate-400 text-xs flex items-center justify-center flex-shrink-0 mt-2 font-mono">
                      {idx + 1}
                    </span>
                    <input
                      type="text"
                      value={b}
                      onChange={(e) => handleBulletChange(idx, e.target.value)}
                      className="flex-1 bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 focus:border-sky-500 outline-none"
                    />
                  </div>
                ))}
              </div>
            </div>

            {/* SEO HTML Description Preview */}
            <div className="space-y-2 pt-2">
              <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                <span className="text-xs font-bold uppercase tracking-wider text-slate-300">
                  SEO Product Description
                </span>
                <div className="flex items-center gap-1 bg-slate-950 p-1 rounded-md border border-slate-800 text-[11px]">
                  <button
                    onClick={() => setActiveTab('copy')}
                    className={`px-2 py-0.5 rounded ${
                      activeTab === 'copy' ? 'bg-slate-800 text-white' : 'text-slate-400'
                    }`}
                  >
                    Formatted
                  </button>
                  <button
                    onClick={() => setActiveTab('html')}
                    className={`px-2 py-0.5 rounded ${
                      activeTab === 'html' ? 'bg-slate-800 text-white' : 'text-slate-400'
                    }`}
                  >
                    HTML Code
                  </button>
                </div>
              </div>

              {activeTab === 'copy' ? (
                <div
                  className="prose prose-invert prose-sm max-w-none p-4 rounded-xl bg-slate-950/80 border border-slate-800 text-xs text-slate-300 max-h-48 overflow-y-auto"
                  dangerouslySetInnerHTML={{ __html: currentProduct.seo_description_html }}
                />
              ) : (
                <textarea
                  readOnly
                  value={currentProduct.seo_description_html}
                  rows={6}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 font-mono text-xs text-slate-400 focus:outline-none"
                />
              )}
            </div>
          </div>
        </div>

        {/* Modal Footer */}
        <div className="px-4 sm:px-6 py-4 border-t border-slate-800 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 bg-slate-950/50">
          <div className="text-xs text-slate-500 truncate max-w-full sm:max-w-md">
            Original Supplier Title: <span className="text-slate-400 italic">{currentProduct.raw_title}</span>
          </div>

          <div className="flex items-center justify-end gap-3 shrink-0 w-full sm:w-auto">
            <button
              onClick={onClose}
              className="px-4 py-2 text-xs font-medium rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 transition shrink-0"
            >
              Cancel
            </button>

            <button
              onClick={handleSave}
              disabled={isSaving}
              className="flex items-center justify-center gap-1.5 px-4 py-2 text-xs font-semibold rounded-lg bg-sky-500 hover:bg-sky-400 text-slate-950 transition shadow-lg shadow-sky-500/20 disabled:opacity-50 whitespace-nowrap shrink-0"
            >
              {isSaving ? <RefreshCw className="w-3.5 h-3.5 animate-spin shrink-0" /> : <Save className="w-3.5 h-3.5 shrink-0" />}
              <span>Save & Recompute Score</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

