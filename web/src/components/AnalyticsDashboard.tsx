import React from 'react';
import { ShieldCheck, AlertTriangle, ShoppingBag, Globe, Zap, BarChart3, ChevronUp, ChevronDown } from 'lucide-react';
import { BatchAnalytics } from '../types';

interface AnalyticsDashboardProps {
  analytics: BatchAnalytics | null;
  isOpen: boolean;
  onToggle: () => void;
  onFilterIssue: (issue: string) => void;
}

export const AnalyticsDashboard: React.FC<AnalyticsDashboardProps> = ({
  analytics,
  isOpen,
  onToggle,
  onFilterIssue,
}) => {
  if (!analytics || analytics.total_products === 0) return null;

  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-2xl overflow-hidden shadow-lg transition duration-200">
      {/* Header bar with toggle */}
      <div
        onClick={onToggle}
        className="px-5 py-3 flex items-center justify-between cursor-pointer hover:bg-slate-800/40 transition select-none"
      >
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-sky-500/10 text-sky-400 flex items-center justify-center border border-sky-500/20">
            <BarChart3 className="w-3.5 h-3.5" />
          </div>
          <div>
            <span className="text-xs font-bold text-slate-100 uppercase tracking-wider">
              Catalog Quality & Compliance Analytics
            </span>
            <span className="text-[11px] text-slate-400 ml-2">
              ({analytics.total_products} SKUs analyzed)
            </span>
          </div>
        </div>

        <div className="flex items-center gap-4 text-xs">
          <div className="flex items-center gap-2">
            <span className="text-slate-400">Compliance Rate:</span>
            <span className="font-mono font-bold text-emerald-400">
              {analytics.compliance_rate.toFixed(1)}%
            </span>
          </div>
          <button className="text-slate-400 hover:text-slate-200 p-1">
            {isOpen ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
        </div>
      </div>

      {/* Expanded Dashboard Body */}
      {isOpen && (
        <div className="p-5 border-t border-slate-800/80 grid grid-cols-1 md:grid-cols-3 gap-6 bg-slate-950/40">
          {/* 1. Overall Compliance Scorecard */}
          <div className="bg-slate-900/60 p-4 rounded-xl border border-slate-800/80 flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                Amazon Pure White Pass
              </span>
              <div className="text-2xl font-bold font-mono text-emerald-400">
                {analytics.compliant_count}{' '}
                <span className="text-xs font-normal text-slate-500">
                  / {analytics.total_products} SKUs
                </span>
              </div>
              <p className="text-[11px] text-slate-400">
                Avg. Listing Score: <span className="text-sky-400 font-semibold">{analytics.average_quality_score.toFixed(1)}%</span>
              </p>
            </div>
            <div className="w-12 h-12 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center">
              <ShieldCheck className="w-6 h-6" />
            </div>
          </div>

          {/* 2. Detected Issues & Bottlenecks */}
          <div className="bg-slate-900/60 p-4 rounded-xl border border-slate-800/80 space-y-2">
            <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block">
              Marketplace Compliance Flags
            </span>
            <div className="flex flex-wrap gap-1.5 text-xs">
              <button
                onClick={() => onFilterIssue('non_white')}
                className="px-2 py-1 rounded-lg bg-amber-500/10 hover:bg-amber-500/20 border border-amber-500/20 text-amber-300 flex items-center gap-1.5 transition text-[11px]"
                title="Filter products failing RGB 255 pure white background"
              >
                <AlertTriangle className="w-3 h-3 text-amber-400" />
                <span>{analytics.non_white_bg_count} Non-White BG</span>
              </button>

              <button
                onClick={() => onFilterIssue('low_res')}
                className="px-2 py-1 rounded-lg bg-rose-500/10 hover:bg-rose-500/20 border border-rose-500/20 text-rose-300 flex items-center gap-1.5 transition text-[11px]"
                title="Filter products with image resolution below 1000px"
              >
                <AlertTriangle className="w-3 h-3 text-rose-400" />
                <span>{analytics.low_resolution_count} Low Res (&lt;1k)</span>
              </button>

              {analytics.watermark_count > 0 && (
                <span className="px-2 py-1 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-300 text-[11px] flex items-center gap-1">
                  {analytics.watermark_count} Overlays
                </span>
              )}
            </div>
          </div>

          {/* 3. Multi-Channel Launch Readiness */}
          <div className="bg-slate-900/60 p-4 rounded-xl border border-slate-800/80 space-y-2">
            <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block">
              Channel Publication Readiness
            </span>
            <div className="grid grid-cols-3 gap-2 text-center text-xs">
              <div className="p-2 rounded-lg bg-slate-950 border border-slate-800">
                <div className="text-[10px] text-slate-500 flex items-center justify-center gap-1">
                  <Zap className="w-2.5 h-2.5 text-amber-400" /> Amazon
                </div>
                <div className="font-mono font-bold text-slate-200 mt-0.5">
                  {analytics.amazon_ready_count}
                </div>
              </div>

              <div className="p-2 rounded-lg bg-slate-950 border border-slate-800">
                <div className="text-[10px] text-slate-500 flex items-center justify-center gap-1">
                  <ShoppingBag className="w-2.5 h-2.5 text-emerald-400" /> Shopify
                </div>
                <div className="font-mono font-bold text-slate-200 mt-0.5">
                  {analytics.shopify_ready_count}
                </div>
              </div>

              <div className="p-2 rounded-lg bg-slate-950 border border-slate-800">
                <div className="text-[10px] text-slate-500 flex items-center justify-center gap-1">
                  <Globe className="w-2.5 h-2.5 text-sky-400" /> Google
                </div>
                <div className="font-mono font-bold text-slate-200 mt-0.5">
                  {analytics.google_ready_count}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
