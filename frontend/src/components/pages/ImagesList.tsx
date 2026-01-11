import { FC, useState, useEffect, useCallback, useMemo } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useStore } from '../../store';
import { SortableTableHeader, useSortState } from '../ui/SortableTableHeader';
import { Pagination } from '../ui/Pagination';
import { formatDate } from '../../lib/utils/formatters';

export const ImagesList: FC = () => {
	const [searchParams, setSearchParams] = useSearchParams();
	const [currentPage, setCurrentPage] = useState(1);
	const itemsPerPage = 50;
	const { sortKey, sortDirection, handleSort } = useSortState('last_scan_date', 'desc');

	// Filters from URL
	const imageFilter = searchParams.get('image') || '';
	const minSeverity = searchParams.get('minSeverity') || '';
	const showUnfixable = searchParams.get('show_unfixable') === 'true'; // Default to false

	const { images, total, loading, error, reload } = useStore((state) => ({
		images: state.images,
		total: state.total,
		loading: state.loading,
		error: state.error,
		reload: state.loadImages
	}));

	// Fetch images when page or filters change
	useEffect(() => {
		reload({
			limit: itemsPerPage,
			offset: (currentPage - 1) * itemsPerPage,
			has_fix: showUnfixable ? undefined : true
		});
	}, [currentPage, showUnfixable, reload]);

	useEffect(() => {
		document.title = 'Images - Invulnerable';
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
	}, [imageFilter, minSeverity, showUnfixable]);

	// Client-side filtering
	const filteredImages = useMemo(() => {
		let filtered = images.filter((image) => {
			// Filter by image name (full name: registry/repository:tag)
			if (imageFilter) {
				const fullImageName = `${image.registry}/${image.repository}:${image.tag}`.toLowerCase();
				if (!fullImageName.includes(imageFilter.toLowerCase())) {
					return false;
				}
			}
			// Filter by minimum severity
			if (minSeverity) {
				switch (minSeverity) {
					case 'Critical':
						if (image.critical_count === 0) return false;
						break;
					case 'High':
						if (image.critical_count === 0 && image.high_count === 0) return false;
						break;
					case 'Medium':
						if (image.critical_count === 0 && image.high_count === 0 && image.medium_count === 0) return false;
						break;
					case 'Low':
						if (image.critical_count === 0 && image.high_count === 0 && image.medium_count === 0 && image.low_count === 0) return false;
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

				if (aVal == null && bVal == null) return 0;
				if (aVal == null) return 1;
				if (bVal == null) return -1;

				if (sortKey === 'last_scan_date') {
					aVal = new Date(aVal).getTime();
					bVal = new Date(bVal).getTime();
				}

				if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
				if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
				return 0;
			});
		}

		return filtered;
	}, [images, imageFilter, minSeverity, sortKey, sortDirection]);

	// Server-side pagination - images already contains only the current page
	const paginatedImages = filteredImages;

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-center">
				<h1 className="text-3xl font-bold text-gray-900">Images</h1>
			</div>

			{/* Filters */}
			<div className="card">
				<h3 className="text-sm font-semibold text-gray-700 mb-3">Filters</h3>
				<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
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
							{total} total images
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
					{(imageFilter || minSeverity) && (
						<button onClick={handleClearFilters} className="btn btn-secondary text-sm">
							Clear Filters
						</button>
					)}
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
							<p className="text-gray-500 text-sm">Loading images...</p>
						</div>
					</div>
				</div>
			) : filteredImages.length === 0 ? (
				<div className="card text-center py-12">
					<p className="text-gray-500">No images found</p>
				</div>
			) : (
				<>
					<div className="card overflow-hidden">
						<div className="overflow-x-auto">
							<table className="min-w-full divide-y divide-gray-200">
								<thead className="bg-gray-50">
									<tr>
										<SortableTableHeader
											label="Image"
											sortKey="repository"
											currentSortKey={sortKey}
											currentSortDirection={sortDirection}
											onSort={handleSort}
										/>
									<SortableTableHeader
										label="Scans"
										sortKey="scan_count"
										currentSortKey={sortKey}
										currentSortDirection={sortDirection}
										onSort={handleSort}
									/>
									<SortableTableHeader
										label="Last Scan"
										sortKey="last_scan_date"
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
								{paginatedImages.map((image) => (
									<tr key={image.id} className="hover:bg-gray-50">
										<td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
											{image.registry}/{image.repository}:{image.tag}
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
											{image.scan_count}
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{image.last_scan_date ? formatDate(image.last_scan_date) : 'Never'}
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm">
											<span className="text-red-600 font-semibold">{image.critical_count}</span>
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm">
											<span className="text-orange-600 font-semibold">{image.high_count}</span>
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm">
											<span className="text-yellow-600 font-semibold">{image.medium_count}</span>
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm">
											<span className="text-blue-600 font-semibold">{image.low_count}</span>
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm">
											<Link
												to={`/images/${image.id}`}
												className="text-blue-600 hover:text-blue-800"
											>
												View History
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
