import { FC, useState, useEffect, useCallback, useMemo } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { api } from '../../lib/api/client';
import type { ImageWithStats } from '../../lib/api/types';
import { formatDate } from '../../lib/utils/formatters';

export const ImagesList: FC = () => {
	const [images, setImages] = useState<ImageWithStats[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [searchParams, setSearchParams] = useSearchParams();

	// Filters from URL
	const registryFilter = searchParams.get('registry') || '';
	const repositoryFilter = searchParams.get('repository') || '';
	const tagFilter = searchParams.get('tag') || '';
	const minVulnCount = searchParams.get('minVulns') || '';
	const showUnfixed = searchParams.get('show_unfixed') === 'true'; // Default to false

	const loadImages = useCallback(
		async (filters?: { registry?: string; repository?: string; tag?: string }) => {
			setLoading(true);
			setError(null);
			try {
				const params: { limit: number; registry?: string; repository?: string; tag?: string; has_fix?: boolean } = { limit: 200 };
				if (filters?.registry) params.registry = filters.registry;
				if (filters?.repository) params.repository = filters.repository;
				if (filters?.tag) params.tag = filters.tag;
				if (!showUnfixed) params.has_fix = true;

				const data = await api.images.list(params);
				setImages(data);
			} catch (e) {
				setError(e instanceof Error ? e.message : 'Failed to load images');
			} finally {
				setLoading(false);
			}
		},
		[showUnfixed]
	);

	useEffect(() => {
		document.title = 'Images - Invulnerable';
		loadImages();
	}, [loadImages]);

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
	const filteredImages = useMemo(() => {
		return images.filter((image) => {
			// Filter by registry
			if (registryFilter && !image.registry.toLowerCase().includes(registryFilter.toLowerCase())) {
				return false;
			}
			// Filter by repository
			if (repositoryFilter && !image.repository.toLowerCase().includes(repositoryFilter.toLowerCase())) {
				return false;
			}
			// Filter by tag
			if (tagFilter && !image.tag.toLowerCase().includes(tagFilter.toLowerCase())) {
				return false;
			}
			// Filter by min vulnerabilities
			const totalVulns =
				image.critical_count + image.high_count + image.medium_count + image.low_count;
			if (minVulnCount && totalVulns < parseInt(minVulnCount)) {
				return false;
			}
			return true;
		});
	}, [images, registryFilter, repositoryFilter, tagFilter, minVulnCount]);

	if (loading) {
		return (
			<div className="text-center py-12">
				<p className="text-gray-500">Loading images...</p>
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
				<h1 className="text-3xl font-bold text-gray-900">Images</h1>
			</div>

			{/* Filters */}
			<div className="card">
				<h3 className="text-sm font-semibold text-gray-700 mb-3">Filters</h3>
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
					<div>
						<label htmlFor="registryFilter" className="block text-sm font-medium text-gray-700">
							Registry
						</label>
						<input
							type="text"
							id="registryFilter"
							value={registryFilter}
							onChange={(e) => updateFilter('registry', e.target.value)}
							placeholder="e.g., docker.io"
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						/>
					</div>

					<div>
						<label
							htmlFor="repositoryFilter"
							className="block text-sm font-medium text-gray-700"
						>
							Repository
						</label>
						<input
							type="text"
							id="repositoryFilter"
							value={repositoryFilter}
							onChange={(e) => updateFilter('repository', e.target.value)}
							placeholder="e.g., library/nginx"
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						/>
					</div>

					<div>
						<label htmlFor="tagFilter" className="block text-sm font-medium text-gray-700">
							Tag
						</label>
						<input
							type="text"
							id="tagFilter"
							value={tagFilter}
							onChange={(e) => updateFilter('tag', e.target.value)}
							placeholder="e.g., latest"
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
							Showing {filteredImages.length} of {images.length} images
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

			{filteredImages.length === 0 ? (
				<div className="card text-center py-12">
					<p className="text-gray-500">No images found</p>
				</div>
			) : (
				<div className="card overflow-hidden">
					<div className="overflow-x-auto">
						<table className="min-w-full divide-y divide-gray-200">
							<thead className="bg-gray-50">
								<tr>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Image
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Scans
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Last Scan
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
								{filteredImages.map((image) => (
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
			)}
		</div>
	);
};
