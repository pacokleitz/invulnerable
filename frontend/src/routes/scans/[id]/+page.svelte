<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import type { ScanWithDetails, Vulnerability } from '$lib/api/types';
	import SeverityBadge from '$lib/components/SeverityBadge.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';

	let scan: ScanWithDetails | null = null;
	let vulnerabilities: Vulnerability[] = [];
	let loading = true;
	let error: string | null = null;

	$: scanId = parseInt($page.params.id);

	onMount(async () => {
		try {
			const data = await api.scans.get(scanId);
			scan = data.scan;
			vulnerabilities = data.vulnerabilities;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load scan';
		} finally {
			loading = false;
		}
	});

	function formatDate(date: string): string {
		return new Date(date).toLocaleString();
	}

	async function viewDiff() {
		window.location.href = `/scans/${scanId}/diff`;
	}

	async function viewSBOM() {
		window.location.href = `/scans/${scanId}/sbom`;
	}
</script>

<svelte:head>
	<title>Scan {scanId} - Invulnerable</title>
</svelte:head>

<div class="space-y-6">
	{#if loading}
		<div class="text-center py-12">
			<p class="text-gray-500">Loading scan...</p>
		</div>
	{:else if error}
		<div class="card bg-red-50">
			<p class="text-red-600">{error}</p>
		</div>
	{:else if scan}
		<div class="flex justify-between items-center">
			<h1 class="text-3xl font-bold text-gray-900">Scan #{scan.id}</h1>
			<div class="space-x-2">
				<button on:click={viewSBOM} class="btn btn-secondary">View SBOM</button>
				<button on:click={viewDiff} class="btn btn-primary">View Diff</button>
			</div>
		</div>

		<!-- Scan Details -->
		<div class="card">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Scan Details</h2>
			<dl class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div>
					<dt class="text-sm font-medium text-gray-500">Image</dt>
					<dd class="mt-1 text-sm text-gray-900">{scan.image_name}</dd>
				</div>
				<div>
					<dt class="text-sm font-medium text-gray-500">Scan Date</dt>
					<dd class="mt-1 text-sm text-gray-900">{formatDate(scan.scan_date)}</dd>
				</div>
				<div>
					<dt class="text-sm font-medium text-gray-500">Grype Version</dt>
					<dd class="mt-1 text-sm text-gray-900">{scan.grype_version || 'N/A'}</dd>
				</div>
				<div>
					<dt class="text-sm font-medium text-gray-500">Total Vulnerabilities</dt>
					<dd class="mt-1 text-sm text-gray-900">{scan.vulnerability_count}</dd>
				</div>
			</dl>
		</div>

		<!-- Severity Summary -->
		<div class="card">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Severity Summary</h2>
			<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
				<div class="flex items-center justify-between p-4 bg-red-50 rounded-lg">
					<div>
						<p class="text-sm font-medium text-red-800">Critical</p>
						<p class="text-2xl font-bold text-red-900">{scan.critical_count}</p>
					</div>
				</div>
				<div class="flex items-center justify-between p-4 bg-orange-50 rounded-lg">
					<div>
						<p class="text-sm font-medium text-orange-800">High</p>
						<p class="text-2xl font-bold text-orange-900">{scan.high_count}</p>
					</div>
				</div>
				<div class="flex items-center justify-between p-4 bg-yellow-50 rounded-lg">
					<div>
						<p class="text-sm font-medium text-yellow-800">Medium</p>
						<p class="text-2xl font-bold text-yellow-900">{scan.medium_count}</p>
					</div>
				</div>
				<div class="flex items-center justify-between p-4 bg-blue-50 rounded-lg">
					<div>
						<p class="text-sm font-medium text-blue-800">Low</p>
						<p class="text-2xl font-bold text-blue-900">{scan.low_count}</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Vulnerabilities List -->
		<div class="card">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Vulnerabilities</h2>
			{#if vulnerabilities.length === 0}
				<p class="text-gray-500">No vulnerabilities found</p>
			{:else}
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">CVE ID</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Package</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Version</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Severity</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Fix Version</th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-gray-200">
							{#each vulnerabilities as vuln}
								<tr class="hover:bg-gray-50">
									<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-blue-600">
										<a href="/vulnerabilities/{vuln.cve_id}" class="hover:underline">{vuln.cve_id}</a>
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{vuln.package_name}</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{vuln.package_version}</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm">
										<SeverityBadge severity={vuln.severity} />
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm">
										<StatusBadge status={vuln.status} />
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
										{vuln.fix_version || 'N/A'}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	{/if}
</div>
