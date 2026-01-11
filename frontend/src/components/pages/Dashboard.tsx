import { FC, useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useStore } from '../../store';
import { SeverityBadge } from '../ui/SeverityBadge';

export const Dashboard: FC = () => {
	const [searchParams, setSearchParams] = useSearchParams();
	const [showUnfixable, setShowUnfixed] = useState(false);
	const { data: metrics, loading, loadMetrics } = useStore((state) => ({
		data: state.data,
		loading: state.loading,
		loadMetrics: state.loadMetrics
	}));

	// Read image filter from URL
	const imageFilter = searchParams.get('image') || '';
	const [imageInput, setImageInput] = useState(imageFilter);

	useEffect(() => {
		document.title = 'Dashboard - Invulnerable';
	}, []);

	useEffect(() => {
		// Sync input with URL parameter
		setImageInput(imageFilter);
	}, [imageFilter]);

	useEffect(() => {
		loadMetrics(showUnfixable ? undefined : true, imageFilter || undefined);
	}, [loadMetrics, showUnfixable, imageFilter]);

	const handleImageFilterChange = (value: string) => {
		setImageInput(value);
	};

	const handleImageFilterSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		const newParams = new URLSearchParams(searchParams);
		if (imageInput.trim()) {
			newParams.set('image', imageInput.trim());
		} else {
			newParams.delete('image');
		}
		setSearchParams(newParams);
	};

	const clearImageFilter = () => {
		setImageInput('');
		const newParams = new URLSearchParams(searchParams);
		newParams.delete('image');
		setSearchParams(newParams);
	};

	if (loading) {
		return (
			<div className="text-center py-12" role="status" aria-live="polite">
				<p className="text-gray-500">Loading metrics...</p>
			</div>
		);
	}

	if (!metrics) {
		return (
			<div className="text-center py-12" role="alert">
				<p className="text-gray-500">No metrics available</p>
			</div>
		);
	}

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-start">
				<div>
					<h1 className="text-3xl font-bold text-gray-900 mb-4" tabIndex={-1}>Dashboard</h1>

					{/* Image Name Filter */}
					<form onSubmit={handleImageFilterSubmit} className="flex items-center gap-2">
						<div className="relative">
							<input
								type="text"
								value={imageInput}
								onChange={(e) => handleImageFilterChange(e.target.value)}
								placeholder="Filter by image name..."
								className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
								style={{ width: '300px' }}
							/>
							{imageFilter && (
								<button
									type="button"
									onClick={clearImageFilter}
									className="absolute right-2 top-2.5 text-gray-400 hover:text-gray-600"
									aria-label="Clear image filter"
								>
									×
								</button>
							)}
						</div>
						<button
							type="submit"
							className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:ring-2 focus:ring-blue-500"
						>
							Apply
						</button>
						{imageFilter && (
							<span className="text-sm text-gray-600">
								Filtering by: <span className="font-medium">{imageFilter}</span>
							</span>
						)}
					</form>
				</div>

				<label className="flex items-center space-x-2 text-sm">
					<input
						type="checkbox"
						checked={showUnfixable}
						onChange={(e) => setShowUnfixed(e.target.checked)}
						className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
					/>
					<span className="text-gray-700">Show unfixable CVEs</span>
				</label>
			</div>

			{/* Summary Cards */}
			<section className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6" aria-label="Summary metrics">
				<article className="card">
					<h3 className="text-sm font-medium text-gray-500">Total Images</h3>
					<p className="mt-2 text-3xl font-bold text-gray-900" aria-label={`${metrics.total_images} total images`}>{metrics.total_images}</p>
				</article>

				<article className="card">
					<h3 className="text-sm font-medium text-gray-500">Total Scans</h3>
					<p className="mt-2 text-3xl font-bold text-gray-900" aria-label={`${metrics.total_scans} total scans`}>{metrics.total_scans}</p>
				</article>

				<article className="card">
					<h3 className="text-sm font-medium text-gray-500">Active Vulnerabilities</h3>
					<p className="mt-2 text-3xl font-bold text-red-600" aria-label={`${metrics.active_vulnerabilities} active vulnerabilities`}>
						{metrics.active_vulnerabilities}
					</p>
				</article>

				<article className="card">
					<h3 className="text-sm font-medium text-gray-500">Scans (24h)</h3>
					<p className="mt-2 text-3xl font-bold text-gray-900" aria-label={`${metrics.recent_scans_24h} scans in last 24 hours`}>{metrics.recent_scans_24h}</p>
				</article>
			</section>

			{/* Severity Breakdown */}
			<section className="card" aria-labelledby="severity-heading">
				<h2 id="severity-heading" className="text-xl font-bold text-gray-900 mb-4">Vulnerabilities by Severity</h2>
				<div className="grid grid-cols-1 md:grid-cols-4 gap-4" role="list">
					<div className="flex items-center justify-between p-4 bg-red-50 rounded-lg">
						<div>
							<p className="text-sm font-medium text-red-800">Critical</p>
							<p className="text-2xl font-bold text-red-900">
								{metrics.severity_counts.critical}
							</p>
						</div>
						<SeverityBadge severity="Critical" />
					</div>

					<div className="flex items-center justify-between p-4 bg-orange-50 rounded-lg">
						<div>
							<p className="text-sm font-medium text-orange-800">High</p>
							<p className="text-2xl font-bold text-orange-900">{metrics.severity_counts.high}</p>
						</div>
						<SeverityBadge severity="High" />
					</div>

					<div className="flex items-center justify-between p-4 bg-yellow-50 rounded-lg">
						<div>
							<p className="text-sm font-medium text-yellow-800">Medium</p>
							<p className="text-2xl font-bold text-yellow-900">
								{metrics.severity_counts.medium}
							</p>
						</div>
						<SeverityBadge severity="Medium" />
					</div>

					<div className="flex items-center justify-between p-4 bg-blue-50 rounded-lg">
						<div>
							<p className="text-sm font-medium text-blue-800">Low</p>
							<p className="text-2xl font-bold text-blue-900">{metrics.severity_counts.low}</p>
						</div>
						<SeverityBadge severity="Low" />
					</div>
				</div>
			</section>
		</div>
	);
};
