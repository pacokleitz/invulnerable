<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import type { ScanWithDetails } from '$lib/api/types';
	import SeverityBadge from '$lib/components/SeverityBadge.svelte';

	let scans: ScanWithDetails[] = [];
	let loading = true;
	let error: string | null = null;

	onMount(async () => {
		try {
			scans = await api.scans.list({ limit: 50 });
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load scans';
		} finally {
			loading = false;
		}
	});

	function formatDate(date: string): string {
		return new Date(date).toLocaleString();
	}
</script>

<svelte:head>
	<title>Scans - Invulnerable</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-gray-900">Scans</h1>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<p class="text-gray-500">Loading scans...</p>
		</div>
	{:else if error}
		<div class="card bg-red-50">
			<p class="text-red-600">{error}</p>
		</div>
	{:else if scans.length === 0}
		<div class="card text-center py-12">
			<p class="text-gray-500">No scans found</p>
		</div>
	{:else}
		<div class="card overflow-hidden">
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Image</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Scan Date</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Total Vulns</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Critical</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">High</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Medium</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Low</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each scans as scan}
							<tr class="hover:bg-gray-50">
								<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{scan.id}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{scan.image_name}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{formatDate(scan.scan_date)}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{scan.vulnerability_count}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<span class="text-red-600 font-semibold">{scan.critical_count}</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<span class="text-orange-600 font-semibold">{scan.high_count}</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<span class="text-yellow-600 font-semibold">{scan.medium_count}</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<span class="text-blue-600 font-semibold">{scan.low_count}</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<a href="/scans/{scan.id}" class="text-blue-600 hover:text-blue-800">View</a>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>
