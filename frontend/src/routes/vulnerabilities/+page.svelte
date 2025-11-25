<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import type { Vulnerability } from '$lib/api/types';
	import SeverityBadge from '$lib/components/SeverityBadge.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';

	let vulnerabilities: Vulnerability[] = [];
	let loading = true;
	let error: string | null = null;

	let severityFilter = '';
	let statusFilter = '';

	onMount(async () => {
		loadVulnerabilities();
	});

	async function loadVulnerabilities() {
		loading = true;
		error = null;

		try {
			const params: any = { limit: 100 };
			if (severityFilter) params.severity = severityFilter;
			if (statusFilter) params.status = statusFilter;

			vulnerabilities = await api.vulnerabilities.list(params);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load vulnerabilities';
		} finally {
			loading = false;
		}
	}

	function formatDate(date: string): string {
		return new Date(date).toLocaleDateString();
	}

	function handleFilterChange() {
		loadVulnerabilities();
	}
</script>

<svelte:head>
	<title>Vulnerabilities - Invulnerable</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-gray-900">Vulnerabilities</h1>
	</div>

	<!-- Filters -->
	<div class="card">
		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			<div>
				<label for="severity" class="block text-sm font-medium text-gray-700">Severity</label>
				<select
					id="severity"
					bind:value={severityFilter}
					on:change={handleFilterChange}
					class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
				>
					<option value="">All</option>
					<option value="Critical">Critical</option>
					<option value="High">High</option>
					<option value="Medium">Medium</option>
					<option value="Low">Low</option>
				</select>
			</div>

			<div>
				<label for="status" class="block text-sm font-medium text-gray-700">Status</label>
				<select
					id="status"
					bind:value={statusFilter}
					on:change={handleFilterChange}
					class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
				>
					<option value="">All</option>
					<option value="active">Active</option>
					<option value="fixed">Fixed</option>
					<option value="ignored">Ignored</option>
					<option value="accepted">Accepted</option>
				</select>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<p class="text-gray-500">Loading vulnerabilities...</p>
		</div>
	{:else if error}
		<div class="card bg-red-50">
			<p class="text-red-600">{error}</p>
		</div>
	{:else if vulnerabilities.length === 0}
		<div class="card text-center py-12">
			<p class="text-gray-500">No vulnerabilities found</p>
		</div>
	{:else}
		<div class="card overflow-hidden">
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">CVE ID</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Package</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Version</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Severity</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">First Detected</th>
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
									{formatDate(vuln.first_detected_at)}
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
</div>
