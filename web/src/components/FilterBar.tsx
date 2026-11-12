import React from 'react';
import { Search, SlidersHorizontal, CheckCircle, AlertTriangle, XCircle } from 'lucide-react';

interface FilterBarProps {
  search: string;
  onSearchChange: (val: string) => void;
  complianceFilter: string;
  onComplianceFilterChange: (val: string) => void;
  statusFilter: string;
  onStatusFilterChange: (val: string) => void;
  totalFiltered: number;
}

export const FilterBar: React.FC<FilterBarProps> = ({
  search,
  onSearchChange,
  complianceFilter,
  onComplianceFilterChange,
  statusFilter,
  onStatusFilterChange,
  totalFiltered,
}) => {
  return (
    <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-3 flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3 shadow-sm">
      {/* Search Input */}
      <div className="relative w-full md:w-80 shrink-0">
        <Search className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
        <input
          type="text"
          placeholder="Search SKU, title, or brand..."
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          className="w-full bg-slate-950 border border-slate-800 rounded-lg pl-9 pr-3 py-1.5 text-xs text-slate-100 placeholder-slate-500 focus:outline-none focus:border-sky-500 transition"
        />
      </div>

      {/* Filter Tabs */}
      <div className="flex items-center flex-wrap gap-2 w-full md:w-auto shrink-0 justify-between md:justify-end">
        <div className="flex items-center gap-1 bg-slate-950 p-1 rounded-lg border border-slate-800 text-xs shrink-0 overflow-x-auto">
          <span className="text-[11px] font-semibold text-slate-500 px-2 flex items-center gap-1 whitespace-nowrap shrink-0">
            <SlidersHorizontal className="w-3 h-3" /> Image:
          </span>
          <button
            onClick={() => onComplianceFilterChange('')}
            className={`px-2.5 py-1 rounded-md transition font-medium whitespace-nowrap shrink-0 ${
              complianceFilter === '' ? 'bg-slate-800 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            All
          </button>
          <button
            onClick={() => onComplianceFilterChange('PASS')}
            className={`px-2.5 py-1 rounded-md flex items-center gap-1 transition font-medium whitespace-nowrap shrink-0 ${
              complianceFilter === 'PASS' ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <CheckCircle className="w-3 h-3 text-emerald-400 shrink-0" />
            Pass
          </button>
          <button
            onClick={() => onComplianceFilterChange('WARN')}
            className={`px-2.5 py-1 rounded-md flex items-center gap-1 transition font-medium whitespace-nowrap shrink-0 ${
              complianceFilter === 'WARN' ? 'bg-amber-500/20 text-amber-400 border border-amber-500/30' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <AlertTriangle className="w-3 h-3 text-amber-400 shrink-0" />
            Warn
          </button>
          <button
            onClick={() => onComplianceFilterChange('FAIL')}
            className={`px-2.5 py-1 rounded-md flex items-center gap-1 transition font-medium whitespace-nowrap shrink-0 ${
              complianceFilter === 'FAIL' ? 'bg-rose-500/20 text-rose-400 border border-rose-500/30' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <XCircle className="w-3 h-3 text-rose-400 shrink-0" />
            Fail
          </button>
        </div>

        {/* Status Filter */}
        <div className="flex items-center gap-1 bg-slate-950 p-1 rounded-lg border border-slate-800 text-xs shrink-0">
          <button
            onClick={() => onStatusFilterChange('')}
            className={`px-2.5 py-1 rounded-md transition font-medium whitespace-nowrap shrink-0 ${
              statusFilter === '' ? 'bg-slate-800 text-white' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            All Status
          </button>
          <button
            onClick={() => onStatusFilterChange('PENDING')}
            className={`px-2.5 py-1 rounded-md transition font-medium whitespace-nowrap shrink-0 ${
              statusFilter === 'PENDING' ? 'bg-sky-500/20 text-sky-400' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Pending
          </button>
          <button
            onClick={() => onStatusFilterChange('APPROVED')}
            className={`px-2.5 py-1 rounded-md transition font-medium whitespace-nowrap shrink-0 ${
              statusFilter === 'APPROVED' ? 'bg-emerald-500/20 text-emerald-400' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Approved
          </button>
        </div>

        <span className="text-xs text-slate-400 font-mono whitespace-nowrap shrink-0">
          {totalFiltered} items
        </span>
      </div>
    </div>
  );
};
