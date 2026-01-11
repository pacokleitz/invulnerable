import { FC, useEffect, useCallback, useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useStore } from '../../store';
import { SortableTableHeader, useSortState } from '../ui/SortableTableHeader';
import { Pagination } from '../ui/Pagination';
import { formatDate } from '../../lib/utils/formatters';

export const ScansList: FC = () => {
	const [searchParams, setSearchParams] = useSearchParams();
	const [currentPage, setCurrentPage] = useState(1);
	const itemsPerPage = 50;

	// Sorting state
	const { sortKey, sortDirection, handleSort } = useSortState('scan_date', 'desc');

	// Filters from URL
	const imageFilter = searchParams.get('image') || '';
	const fromDate = searchParams.get('from') || '';
	const toDate = searchParams.get('to') || '';
	const minSeverity = searchParams.get('minSeverity') || '';
	const showUnfixable = searchParams.get('show_unfixable') === 'true'; // Default to false

	const { scans, total, loading, error, reload } = useStore((state) => ({
		scans: state.scans,
		total: state.total,
		loading: state.loading,
		error: state.error,
		reload: state.loadScans
	}));

	// Fetch scans when page or filters change
	useEffect(() => {
		reload({
			limit: itemsPerPage,
			offset: (currentPage - 1) * itemsPerPage,
			image: imageFilter || undefined,
			has_fix: showUnfixable ? undefined : true
		});
	}, [currentPage, imageFilter, showUnfixable, reload]);

	useEffect(() => {
		document.title = 'Scans - Invulnerable';
	}, []);

	const updateFilter = useCallback((key: string, value: string) => {
		setSearchParams(prev => {
			const newParams = new URLSearchParams(prev);
			if (value) {
				newParams.set(key, value);
			} else {
				newParams.delete(key);
			}
			return newParams;
		});
	}, [setSearchParams]);

	const handleClearFilters = useCallback(() => {
		setSearchParams({});
		setCurrentPage(1);
	}, [setSearchParams]);

	// Reset to page 1 when filters change
	useEffect(() => {
		setCurrentPage(1);
	}, [imageFilter, fromDate, toDate, minSeverity, showUnfixable]);

	// Apply client-side filters that are not yet supported by backend (date filters and minSeverity)
	// Image filter is now handled server-side
	const filteredScans = useMemo(() => {
		let filtered = scans.filter((scan) => {
			// Filter by from date
			if (fromDate && new Date(scan.scan_date) < new Date(fromDate)) {
				return false;
			}
			// Filter by to date
			if (toDate && new Date(scan.scan_date) > new Date(toDate)) {
				return false;
			}
			// Filter by minimum severity
			if (minSeverity) {
				switch (minSeverity) {
					case 'Critical':
						if (scan.critical_count === 0) return false;
						break;
					case 'High':
						if (scan.critical_count === 0 && scan.high_count === 0) return false;
						break;
					case 'Medium':
						if (scan.critical_count === 0 && scan.high_count === 0 && scan.medium_count === 0) return false;
						break;
					case 'Low':
						if (scan.critical_count === 0 && scan.high_count === 0 && scan.medium_count === 0 && scan.low_count === 0) return false;
						break;
				}
			}
			return true;
		});

		// Apply sorting
		if (sortKey && sortDirection) {
			filtered.sort((a, b) => {
				let aVal: any = a[sortKey as keyof typeof a];
				let bVal: any = b[sortKey as keyof typeof b];

				// Handle null/undefined
				if (aVal == null && bVal == null) return 0;
				if (aVal == null) return 1;
				if (bVal == null) return -1;

				// Handle dates
				if (sortKey === 'scan_date') {
					aVal = new Date(aVal).getTime();
					bVal = new Date(bVal).getTime();
				}

				// Compare
				if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
				if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
				return 0;
			});
		}

		return filtered;
	}, [scans, fromDate, toDate, minSeverity, sortKey, sortDirection]);

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-center">
				<h1 className="text-3xl font-bold text-gray-900">Scans</h1>
			</div>

			{/* Filters */}
			<div className="card">
				<h3 className="text-sm font-semibold text-gray-700 mb-3">Filters</h3>
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
					<div>
						<label htmlFor="imageFilter" className="block text-sm font-medium text-gray-700">
							Image Name
						</label>
						<input
							type="text"
							id="imageFilter"
							value={imageFilter}
							onChange={(e) => updateFilter('image', e.target.value)}
							placeholder="e.g., nginx:latest"
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						/>
					</div>

					<div>
						<label htmlFor="fromDate" className="block text-sm font-medium text-gray-700">
							From Date
						</label>
						<input
							type="date"
							id="fromDate"
							value={fromDate}
							onChange={(e) => updateFilter('from', e.target.value)}
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						/>
					</div>

					<div>
						<label htmlFor="toDate" className="block text-sm font-medium text-gray-700">
							To Date
						</label>
						<input
							type="date"
							id="toDate"
							value={toDate}
							onChange={(e) => updateFilter('to', e.target.value)}
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						/>
					</div>

					<div>
						<label htmlFor="minSeverity" className="block text-sm font-medium text-gray-700">
							Severity
						</label>
						<select
							id="minSeverity"
							value={minSeverity}
							onChange={(e) => updateFilter('minSeverity', e.target.value)}
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						>
							<option value="">All</option>
							<option value="Critical">Critical</option>
							<option value="High">High or higher</option>
							<option value="Medium">Medium or higher</option>
							<option value="Low">Low or higher</option>
						</select>
					</div>
				</div>

				<div className="mt-4 flex justify-between items-center">
					<div className="flex items-center space-x-4">
						<p className="text-sm text-gray-600">
							{total} total scans
						</p>
						<label className="flex items-center space-x-2 text-sm">
							<input
								type="checkbox"
								checked={showUnfixable}
								onChange={(e) => updateFilter('show_unfixable', e.target.checked ? 'true' : 'false')}
								className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
							/>
							<span className="text-gray-700">Show unfixable CVEs</span>
						</label>
					</div>
					<button onClick={handleClearFilters} className="btn btn-secondary text-sm">
						Clear Filters
					</button>
				</div>
			</div>

			{error && (
				<div className="card bg-red-50">
					<p className="text-red-600">{error}</p>
				</div>
			)}

			{loading ? (
				<div className="card">
					<div className="flex items-center justify-center py-32">
						<div className="flex flex-col items-center gap-3">
							<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
							<p className="text-gray-500 text-sm">Loading scans...</p>
						</div>
					</div>
				</div>
			) : filteredScans.length === 0 ? (
				<div className="card text-center py-12">
					<p className="text-gray-500">No scans found</p>
				</div>
			) : (
				<>
					<div className="card overflow-hidden">
						<div className="overflow-x-auto">
							<table className="min-w-full divide-y divide-gray-200">
								<thead className="bg-gray-50">
									<tr>
										<SortableTableHeader
											label="ID"
											sortKey="id"
											currentSortKey={sortKey}
											currentSortDirection={sortDirection}
											onSort={handleSort}
										/>
										<SortableTableHeader
											label="Image"
											sortKey="image_name"
											currentSortKey={sortKey}
											currentSortDirection={sortDirection}
											onSort={handleSort}
										/>
										<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
											Image Digest
										</th>
										<SortableTableHeader
											label="Scan Date"
											sortKey="scan_date"
											currentSortKey={sortKey}
											currentSortDirection={sortDirection}
											onSort={handleSort}
										/>
										<SortableTableHeader
											label="Total Vulns"
											sortKey="vulnerability_count"
											currentSortKey={sortKey}
											currentSortDirection={sortDirection}
											onSort={handleSort}
										/>
										<SortableTableHeader
											label="Critical"
											sortKey="critical_count"
											currentSortKey={sortKey}
											currentSortDirection={sortDirection}
											onSort={handleSort}
										/>
										<SortableTableHeader
											label="High"
											sortKey="high_count"
											currentSortKey={sortKey}
											currentSortDirection={sortDirection}
											onSort={handleSort}
										/>
										<SortableTableHeader
											label="Medium"
											sortKey="medium_count"
											currentSortKey={sortKey}
											currentSortDirection={sortDirection}
											onSort={handleSort}
										/>
										<SortableTableHeader
											label="Low"
											sortKey="low_count"
											currentSortKey={sortKey}
											currentSortDirection={sortDirection}
											onSort={handleSort}
										/>
										<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
											Actions
										</th>
									</tr>
								</thead>
								<tbody className="bg-white divide-y divide-gray-200">
									{filteredScans.map((scan) => (
										<tr key={scan.id} className="hover:bg-gray-50">
											<td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
												{scan.id}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
												{scan.image_name}
											</td>
											<td
												className="px-6 py-4 text-sm text-gray-500 font-mono max-w-xs truncate"
												title={scan.image_digest || 'N/A'}
											>
												{scan.image_digest || 'N/A'}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
												{formatDate(scan.scan_date)}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
												{scan.vulnerability_count}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<span className="text-red-600 font-semibold">{scan.critical_count}</span>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<span className="text-orange-600 font-semibold">{scan.high_count}</span>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<span className="text-yellow-600 font-semibold">{scan.medium_count}</span>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<span className="text-blue-600 font-semibold">{scan.low_count}</span>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<Link to={`/scans/${scan.id}`} className="text-blue-600 hover:text-blue-800">
													View
												</Link>
											</td>
										</tr>
									))}
								</tbody>
							</table>
						</div>
					</div>

					<Pagination
						currentPage={currentPage}
						totalItems={total}
						itemsPerPage={itemsPerPage}
						onPageChange={setCurrentPage}
					/>
				</>
			)}
		</div>
	);
};
