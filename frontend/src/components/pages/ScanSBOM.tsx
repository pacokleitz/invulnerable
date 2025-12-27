import { FC, useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../../lib/api/client';

export const ScanSBOM: FC = () => {
	const { id } = useParams<{ id: string }>();
	const scanId = parseInt(id || '0', 10);
	const [sbom, setSbom] = useState<Record<string, unknown> | null>(null);
	const [sbomJson, setSbomJson] = useState('');
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [copied, setCopied] = useState(false);

	useEffect(() => {
		document.title = `SBOM - Scan ${scanId} - Invulnerable`;

		const fetchSBOM = async () => {
			try {
				const data = await api.scans.getSBOM(scanId);
				setSbom(data);
				setSbomJson(JSON.stringify(data, null, 2));
			} catch (e) {
				setError(e instanceof Error ? e.message : 'Failed to load SBOM');
			} finally {
				setLoading(false);
			}
		};

		fetchSBOM();
	}, [scanId]);

	const downloadSBOM = useCallback(() => {
		const blob = new Blob([sbomJson], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `scan-${scanId}-sbom.json`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}, [sbomJson, scanId]);

	const copySBOM = useCallback(() => {
		// Try modern clipboard API first
		if (navigator.clipboard && navigator.clipboard.writeText) {
			navigator.clipboard
				.writeText(sbomJson)
				.then(() => {
					setCopied(true);
					setTimeout(() => {
						setCopied(false);
					}, 2000);
				})
				.catch(() => {
					// Fallback to legacy method
					fallbackCopyToClipboard(sbomJson);
				});
		} else {
			// Fallback for browsers without clipboard API
			fallbackCopyToClipboard(sbomJson);
		}
	}, [sbomJson]);

	const fallbackCopyToClipboard = (text: string) => {
		const textArea = document.createElement('textarea');
		textArea.value = text;
		textArea.style.position = 'fixed';
		textArea.style.left = '-999999px';
		textArea.style.top = '-999999px';
		document.body.appendChild(textArea);
		textArea.focus();
		textArea.select();

		try {
			document.execCommand('copy');
			setCopied(true);
			setTimeout(() => {
				setCopied(false);
			}, 2000);
		} catch (err) {
			console.error('Failed to copy:', err);
			alert('Failed to copy to clipboard. Please copy manually.');
		} finally {
			document.body.removeChild(textArea);
		}
	};

	if (loading) {
		return (
			<div className="text-center py-12">
				<p className="text-gray-500">Loading SBOM...</p>
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

	if (!sbom) {
		return null;
	}

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-center">
				<h1 className="text-3xl font-bold text-gray-900">Software Bill of Materials</h1>
				<Link to={`/scans/${scanId}`} className="text-blue-600 hover:text-blue-800">
					← Back to Scan
				</Link>
			</div>

			<div className="card">
				<div className="flex justify-between items-center mb-4">
					<h2 className="text-xl font-semibold text-gray-900">Scan #{scanId}</h2>
					<div className="space-x-2">
						<button onClick={copySBOM} className="btn btn-secondary">
							{copied ? 'Copied!' : 'Copy to Clipboard'}
						</button>
						<button onClick={downloadSBOM} className="btn btn-primary">
							Download JSON
						</button>
					</div>
				</div>

				<div className="bg-gray-900 rounded-lg p-4 overflow-auto max-h-[600px]">
					<pre className="text-sm text-gray-100">
						<code>{sbomJson}</code>
					</pre>
				</div>
			</div>
		</div>
	);
};
