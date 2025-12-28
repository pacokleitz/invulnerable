import { FC, useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useImageHistory } from '../../hooks/useImages';
import { formatDate } from '../../lib/utils/formatters';

export const ImageHistory: FC = () => {
	const { id } = useParams<{ id: string }>();
	const imageId = parseInt(id || '0', 10);
	const [showUnfixed, setShowUnfixed] = useState(false);
	const { currentImageHistory, loading, error } = useImageHistory(imageId, 50, showUnfixed ? undefined : true);

	useEffect(() => {
		document.title = 'Image History - Invulnerable';
	}, []);

	if (loading) {
		return (
			<div className="text-center py-12">
				<p className="text-gray-500">Loading scan history...</p>
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
				<h1 className="text-3xl font-bold text-gray-900">Image Scan History</h1>
				<Link to="/images" className="text-blue-600 hover:text-blue-800">
					← Back to Images
				</Link>
			</div>

			{currentImageHistory.length === 0 ? (
				<div className="card text-center py-12">
					<p className="text-gray-500">No scan history found for this image</p>
				</div>
			) : (
				<>
					<div className="card">
						<div className="flex justify-between items-center mb-4">
							<div>
								<h2 className="text-xl font-semibold text-gray-900">
									{currentImageHistory[0].image_name}
								</h2>
								<p className="text-sm text-gray-600 mt-2">
									Total scans: {currentImageHistory.length}
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
										<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
											ID
										</th>
										<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
											Scan Date
										</th>
										<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
											Image Digest
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
									{currentImageHistory.map((scan) => (
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
				</>
			)}
		</div>
	);
};
