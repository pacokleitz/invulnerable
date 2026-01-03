import { FC, useEffect, useState, useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useStore } from '../../store';
import { SortableTableHeader, useSortState } from '../ui/SortableTableHeader';
import { Pagination } from '../ui/Pagination';
import { formatDate } from '../../lib/utils/formatters';
import type { ScanWithDetails } from '../../lib/api/types';

export const ImageHistory: FC = () => {
	const { id } = useParams<{ id: string }>();
	const imageId = parseInt(id || '0', 10);
	const [currentPage, setCurrentPage] = useState(1);
	const itemsPerPage = 50;
	const [showUnfixed, setShowUnfixed] = useState(false);
	const { sortKey, sortDirection, handleSort } = useSortState('scan_date', 'desc');

	const { currentImageHistory, historyTotal, loading, error, reload } = useStore((state) => ({
		currentImageHistory: state.currentImageHistory,
		historyTotal: state.historyTotal,
		loading: state.loading,
		error: state.error,
		reload: state.loadImageHistory
	}));

	// Fetch scans when page or filters change
	useEffect(() => {
		reload(imageId, itemsPerPage, (currentPage - 1) * itemsPerPage, showUnfixed ? undefined : true);
	}, [imageId, currentPage, showUnfixed, reload]);

	useEffect(() => {
		document.title = 'Image History - Invulnerable';
	}, []);

	// Reset to page 1 when filters change
	useEffect(() => {
		setCurrentPage(1);
	}, [showUnfixed]);

	// Apply sorting to scan history
	const sortedHistory = useMemo(() => {
		if (!sortKey || !sortDirection) {
			return currentImageHistory;
		}

		const sorted = [...currentImageHistory].sort((a, b) => {
			let aVal: any = a[sortKey as keyof ScanWithDetails];
			let bVal: any = b[sortKey as keyof ScanWithDetails];

			// Handle null/undefined values
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

		return sorted;
	}, [currentImageHistory, sortKey, sortDirection]);

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-center">
				<h1 className="text-3xl font-bold text-gray-900">Image Scan History</h1>
				<Link to="/images" className="text-blue-600 hover:text-blue-800">
					← Back to Images
				</Link>
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
							<p className="text-gray-500 text-sm">Loading scan history...</p>
						</div>
					</div>
				</div>
			) : sortedHistory.length === 0 ? (
				<div className="card text-center py-12">
					<p className="text-gray-500">No scan history found for this image</p>
				</div>
			) : (
				<>
					<div className="card">
						<div className="flex justify-between items-center mb-4">
							<div>
								<h2 className="text-xl font-semibold text-gray-900">
									{sortedHistory[0]?.image_name || 'Unknown'}
								</h2>
								<p className="text-sm text-gray-600 mt-2">
									Total scans: {historyTotal}
								</p>
							</div>
							<label className="flex items-center space-x-2 text-sm">
								<input
									type="checkbox"
									checked={showUnfixed}
									onChange={(e) => setShowUnfixed(e.target.checked)}
									className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
								/>
								<span className="text-gray-700">Show unfixed CVEs</span>
							</label>
						</div>
					</div>

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
											label="Scan Date"
											sortKey="scan_date"
											currentSortKey={sortKey}
											currentSortDirection={sortDirection}
											onSort={handleSort}
										/>
										<SortableTableHeader
											label="Image Digest"
											sortKey="image_digest"
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
									{sortedHistory.map((scan) => (
										<tr key={scan.id} className="hover:bg-gray-50">
											<td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
												{scan.id}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
												{formatDate(scan.scan_date)}
											</td>
											<td
												className="px-6 py-4 text-sm text-gray-500 font-mono max-w-xs truncate"
												title={scan.image_digest || 'N/A'}
											>
												{scan.image_digest || 'N/A'}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
												{scan.vulnerability_count}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<span className="text-red-600 font-semibold">
													{scan.critical_count}
												</span>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<span className="text-orange-600 font-semibold">{scan.high_count}</span>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<span className="text-yellow-600 font-semibold">
													{scan.medium_count}
												</span>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<span className="text-blue-600 font-semibold">{scan.low_count}</span>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<Link
													to={`/scans/${scan.id}`}
													className="text-blue-600 hover:text-blue-800"
												>
													View Details
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
						totalItems={historyTotal}
						itemsPerPage={itemsPerPage}
						onPageChange={setCurrentPage}
					/>
				</>
			)}
		</div>
	);
};
