import { FC, useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../../lib/api/client';
import type { Vulnerability } from '../../lib/api/types';
import { SeverityBadge } from '../ui/SeverityBadge';
import { StatusBadge } from '../ui/StatusBadge';
import { formatDate } from '../../lib/utils/formatters';

export const CVEDetails: FC = () => {
	const { cve } = useParams<{ cve: string }>();
	const [vulnerabilities, setVulnerabilities] = useState<Vulnerability[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [showUpdateModal, setShowUpdateModal] = useState(false);
	const [selectedVuln, setSelectedVuln] = useState<Vulnerability | null>(null);
	const [updateStatus, setUpdateStatus] = useState('');
	const [updateNotes, setUpdateNotes] = useState('');
	const [updating, setUpdating] = useState(false);

	const loadVulnerabilities = useCallback(async () => {
		if (!cve) return;
		setLoading(true);
		setError(null);
		try {
			const data = await api.vulnerabilities.getByCVE(cve);
			setVulnerabilities(data);
		} catch (e) {
			setError(e instanceof Error ? e.message : 'Failed to load vulnerability');
		} finally {
			setLoading(false);
		}
	}, [cve]);

	useEffect(() => {
		document.title = `${cve} - Invulnerable`;
		loadVulnerabilities();
	}, [cve, loadVulnerabilities]);

	const openUpdateModal = useCallback((vuln: Vulnerability) => {
		setSelectedVuln(vuln);
		setUpdateStatus(vuln.status);
		setUpdateNotes(vuln.notes || '');
		setShowUpdateModal(true);
	}, []);

	const closeUpdateModal = useCallback(() => {
		setShowUpdateModal(false);
		setSelectedVuln(null);
		setUpdateStatus('');
		setUpdateNotes('');
	}, []);

	const handleUpdate = useCallback(async () => {
		if (!selectedVuln) return;

		setUpdating(true);
		try {
			await api.vulnerabilities.update(selectedVuln.id, {
				status: updateStatus,
				notes: updateNotes || undefined
			});
			await loadVulnerabilities();
			closeUpdateModal();
		} catch (e) {
			alert(e instanceof Error ? e.message : 'Failed to update vulnerability');
		} finally {
			setUpdating(false);
		}
	}, [selectedVuln, updateStatus, updateNotes, loadVulnerabilities, closeUpdateModal]);

	if (loading) {
		return (
			<div className="text-center py-12">
				<p className="text-gray-500">Loading vulnerability details...</p>
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

	if (vulnerabilities.length === 0) {
		return (
			<div className="card text-center py-12">
				<p className="text-gray-500">Vulnerability not found</p>
			</div>
		);
	}

	const cveInfo = vulnerabilities[0];

	return (
		<>
			<div className="space-y-6">
				<div className="flex justify-between items-center">
					<h1 className="text-3xl font-bold text-gray-900">{cve}</h1>
					<Link to="/vulnerabilities" className="text-blue-600 hover:text-blue-800">
						← Back to Vulnerabilities
					</Link>
				</div>

				{/* CVE Summary */}
				<div className="card">
					<h2 className="text-xl font-bold text-gray-900 mb-4">Vulnerability Details</h2>
					<dl className="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<dt className="text-sm font-medium text-gray-500">CVE ID</dt>
							<dd className="mt-1 text-sm text-gray-900">{cveInfo.cve_id}</dd>
						</div>
						<div>
							<dt className="text-sm font-medium text-gray-500">Severity</dt>
							<dd className="mt-1">
								<SeverityBadge severity={cveInfo.severity} />
							</dd>
						</div>
						<div>
							<dt className="text-sm font-medium text-gray-500">Affected Packages</dt>
							<dd className="mt-1 text-sm text-gray-900">{vulnerabilities.length}</dd>
						</div>
					</dl>
				</div>

				{/* Affected Packages */}
				<div className="card">
					<h2 className="text-xl font-bold text-gray-900 mb-4">Affected Packages</h2>
					<div className="overflow-x-auto">
						<table className="min-w-full divide-y divide-gray-200">
							<thead className="bg-gray-50">
								<tr>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Package
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Version
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Status
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										First Detected
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Fix Version
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Actions
									</th>
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{vulnerabilities.map((vuln) => (
									<tr key={vuln.id} className="hover:bg-gray-50">
										<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
											{vuln.package_name}
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{vuln.package_version}
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm">
											<StatusBadge status={vuln.status} />
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{formatDate(vuln.first_detected_at)}
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{vuln.fix_version || 'N/A'}
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm">
											<button
												onClick={() => openUpdateModal(vuln)}
												className="text-blue-600 hover:text-blue-800"
											>
												Update
											</button>
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				</div>
			</div>

			{/* Update Modal */}
			{showUpdateModal && (
				<div
					className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50"
					role="dialog"
					aria-modal="true"
					aria-labelledby="modal-title"
				>
					<div className="bg-white rounded-lg max-w-md w-full p-6">
						<h3 id="modal-title" className="text-lg font-bold text-gray-900 mb-4">Update Vulnerability</h3>
						<div className="space-y-4">
							<div>
								<label htmlFor="status-select" className="block text-sm font-medium text-gray-700">Status</label>
								<select
									id="status-select"
									value={updateStatus}
									onChange={(e) => setUpdateStatus(e.target.value)}
									className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
									aria-label="Vulnerability status"
								>
									<option value="active">Active</option>
									<option value="fixed">Fixed</option>
									<option value="ignored">Ignored</option>
									<option value="accepted">Accepted</option>
								</select>
							</div>
							<div>
								<label htmlFor="notes-textarea" className="block text-sm font-medium text-gray-700">Notes</label>
								<textarea
									id="notes-textarea"
									value={updateNotes}
									onChange={(e) => setUpdateNotes(e.target.value)}
									rows={3}
									className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
									aria-label="Additional notes"
								/>
							</div>
						</div>
						<div className="mt-6 flex justify-end space-x-2">
							<button onClick={closeUpdateModal} className="btn btn-secondary" aria-label="Cancel update">
								Cancel
							</button>
							<button
								onClick={handleUpdate}
								disabled={updating}
								className="btn btn-primary"
								aria-label={updating ? 'Updating vulnerability' : 'Update vulnerability'}
							>
								{updating ? 'Updating...' : 'Update'}
							</button>
						</div>
					</div>
				</div>
			)}
		</>
	);
};
