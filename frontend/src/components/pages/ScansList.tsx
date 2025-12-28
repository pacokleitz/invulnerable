import { FC, useEffect, useCallback, useMemo } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useScans } from '../../hooks/useScans';
import { formatDate } from '../../lib/utils/formatters';

export const ScansList: FC = () => {
	const [searchParams, setSearchParams] = useSearchParams();

	// Filters from URL
	const imageFilter = searchParams.get('image') || '';
	const fromDate = searchParams.get('from') || '';
	const toDate = searchParams.get('to') || '';
	const minVulnCount = searchParams.get('minVulns') || '';
	const showUnfixed = searchParams.get('show_unfixed') === 'true'; // Default to false

	const { scans, loading, error } = useScans({
		limit: 100,
		has_fix: showUnfixed ? undefined : true
	});

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
	}, [setSearchParams]);

	// Client-side filtering
	const filteredScans = useMemo(() => {
		return scans.filter((scan) => {
			// Filter by image name
			if (imageFilter && !scan.image_name.toLowerCase().includes(imageFilter.toLowerCase())) {
				return false;
			}
			// Filter by from date
			if (fromDate && new Date(scan.scan_date) < new Date(fromDate)) {
				return false;
			}
			// Filter by to date
			if (toDate && new Date(scan.scan_date) > new Date(toDate)) {
				return false;
			}
			// Filter by min vulnerabilities
			if (minVulnCount && scan.vulnerability_count < parseInt(minVulnCount)) {
				return false;
			}
			return true;
		});
	}, [scans, imageFilter, fromDate, toDate, minVulnCount]);

	if (loading) {
		return (
			<div className="text-center py-12">
				<p className="text-gray-500">Loading scans...</p>
			</div>
		);
	}

	if (error) {
		return (
			<div className="card bg-red-50">
				<p className="text-red-600">{error}</p>
			</div>
		);
	}

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
							placeholder="Search by image..."
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
						<label htmlFor="minVulnCount" className="block text-sm font-medium text-gray-700">
							Min Vulnerabilities
						</label>
						<input
							type="number"
							id="minVulnCount"
							value={minVulnCount}
							onChange={(e) => updateFilter('minVulns', e.target.value)}
							placeholder="0"
							min="0"
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						/>
					</div>
				</div>

				<div className="mt-4 flex justify-between items-center">
					<div className="flex items-center space-x-4">
						<p className="text-sm text-gray-600">
							Showing {filteredScans.length} of {scans.length} scans
						</p>
						<label className="flex items-center space-x-2 text-sm">
							<input
								type="checkbox"
								checked={showUnfixed}
								onChange={(e) => updateFilter('show_unfixed', e.target.checked ? 'true' : 'false')}
								className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
							/>
							<span className="text-gray-700">Show unfixed CVEs</span>
						</label>
					</div>
					<button onClick={handleClearFilters} className="btn btn-secondary text-sm">
						Clear Filters
					</button>
				</div>
			</div>

			{filteredScans.length === 0 ? (
				<div className="card text-center py-12">
					<p className="text-gray-500">No scans found</p>
				</div>
			) : (
				<div className="card overflow-hidden">
					<div className="overflow-x-auto">
						<table className="min-w-full divide-y divide-gray-200">
							<thead className="bg-gray-50">
								<tr>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										ID
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Image
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Image Digest
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Scan Date
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Total Vulns
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Critical
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										High
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Medium
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Low
									</th>
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
			)}
		</div>
	);
};
