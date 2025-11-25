<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import type { ScanDiff } from '$lib/api/types';
	import SeverityBadge from '$lib/components/SeverityBadge.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';

	let diff: ScanDiff | null = null;
	let loading = true;
	let error: string | null = null;

	$: scanId = parseInt($page.params.id);

	onMount(async () => {
		try {
			diff = await api.scans.getDiff(scanId);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load scan diff';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Scan Diff - Scan {scanId} - Invulnerable</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-gray-900">Scan Comparison</h1>
		<a href="/scans/{scanId}" class="text-blue-600 hover:text-blue-800">← Back to Scan</a>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<p class="text-gray-500">Loading scan comparison...</p>
		</div>
	{:else if error}
		<div class="card bg-red-50">
			<p class="text-red-600">{error}</p>
		</div>
	{:else if diff}
		<!-- Summary Cards -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
			<div class="card bg-green-50">
				<h3 class="text-sm font-medium text-green-800">Fixed Vulnerabilities</h3>
				<p class="mt-2 text-3xl font-bold text-green-900">{diff.summary.fixed_count}</p>
				<p class="mt-1 text-xs text-green-700">Resolved since previous scan</p>
			</div>

			<div class="card bg-red-50">
				<h3 class="text-sm font-medium text-red-800">New Vulnerabilities</h3>
				<p class="mt-2 text-3xl font-bold text-red-900">{diff.summary.new_count}</p>
				<p class="mt-1 text-xs text-red-700">Introduced since previous scan</p>
			</div>

			<div class="card bg-gray-50">
				<h3 class="text-sm font-medium text-gray-800">Persistent Vulnerabilities</h3>
				<p class="mt-2 text-3xl font-bold text-gray-900">{diff.summary.persistent_count}</p>
				<p class="mt-1 text-xs text-gray-700">Still present from previous scan</p>
			</div>
		</div>

		<div class="card">
			<p class="text-sm text-gray-600">
				Comparing Scan #{diff.scan_id} with Scan #{diff.previous_scan_id}
			</p>
		</div>

		<!-- New Vulnerabilities -->
		{#if diff.new_vulnerabilities.length > 0}
			<div class="card">
				<h2 class="text-xl font-bold text-red-900 mb-4">
					New Vulnerabilities ({diff.new_vulnerabilities.length})
				</h2>
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">CVE ID</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Package</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Version</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Severity</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Fix Version</th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-gray-200">
							{#each diff.new_vulnerabilities as vuln}
								<tr class="hover:bg-gray-50">
									<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-blue-600">
										<a href="/vulnerabilities/{vuln.cve_id}" class="hover:underline">{vuln.cve_id}</a>
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{vuln.package_name}</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{vuln.package_version}</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm">
										<SeverityBadge severity={vuln.severity} />
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
										{vuln.fix_version || 'N/A'}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}

		<!-- Fixed Vulnerabilities -->
		{#if diff.fixed_vulnerabilities.length > 0}
			<div class="card">
				<h2 class="text-xl font-bold text-green-900 mb-4">
					Fixed Vulnerabilities ({diff.fixed_vulnerabilities.length})
				</h2>
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">CVE ID</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Package</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Version</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Severity</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Fix Version</th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-gray-200">
							{#each diff.fixed_vulnerabilities as vuln}
								<tr class="hover:bg-gray-50">
									<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-blue-600">
										<a href="/vulnerabilities/{vuln.cve_id}" class="hover:underline">{vuln.cve_id}</a>
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{vuln.package_name}</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{vuln.package_version}</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm">
										<SeverityBadge severity={vuln.severity} />
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
										{vuln.fix_version || 'N/A'}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}

		<!-- Persistent Vulnerabilities -->
		{#if diff.persistent_vulnerabilities.length > 0}
			<div class="card">
				<h2 class="text-xl font-bold text-gray-900 mb-4">
					Persistent Vulnerabilities ({diff.persistent_vulnerabilities.length})
				</h2>
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
							{#each diff.persistent_vulnerabilities as vuln}
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
			</div>
		{/if}

		<!-- Empty state if no vulnerabilities -->
		{#if diff.summary.new_count === 0 && diff.summary.fixed_count === 0 && diff.summary.persistent_count === 0}
			<div class="card text-center py-12">
				<p class="text-gray-500">No vulnerabilities to compare</p>
			</div>
		{/if}
	{/if}
</div>
