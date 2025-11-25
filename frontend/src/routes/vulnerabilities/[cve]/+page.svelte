<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import type { Vulnerability } from '$lib/api/types';
	import SeverityBadge from '$lib/components/SeverityBadge.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';

	let vulnerabilities: Vulnerability[] = [];
	let loading = true;
	let error: string | null = null;
	let selectedVuln: Vulnerability | null = null;
	let showUpdateModal = false;
	let updateStatus = '';
	let updateNotes = '';
	let updating = false;

	$: cveId = $page.params.cve;

	onMount(async () => {
		await loadVulnerabilities();
	});

	async function loadVulnerabilities() {
		loading = true;
		error = null;
		try {
			vulnerabilities = await api.vulnerabilities.getByCVE(cveId);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load vulnerability';
		} finally {
			loading = false;
		}
	}

	function formatDate(date: string): string {
		return new Date(date).toLocaleString();
	}

	function openUpdateModal(vuln: Vulnerability) {
		selectedVuln = vuln;
		updateStatus = vuln.status;
		updateNotes = vuln.notes || '';
		showUpdateModal = true;
	}

	function closeUpdateModal() {
		showUpdateModal = false;
		selectedVuln = null;
		updateStatus = '';
		updateNotes = '';
	}

	async function updateVulnerability() {
		if (!selectedVuln) return;

		updating = true;
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
			updating = false;
		}
	}

	// Get unique CVE information from first vulnerability
	$: cveInfo = vulnerabilities.length > 0 ? vulnerabilities[0] : null;
</script>

<svelte:head>
	<title>{cveId} - Invulnerable</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-gray-900">{cveId}</h1>
		<a href="/vulnerabilities" class="text-blue-600 hover:text-blue-800">← Back to Vulnerabilities</a>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<p class="text-gray-500">Loading vulnerability details...</p>
		</div>
	{:else if error}
		<div class="card bg-red-50">
			<p class="text-red-600">{error}</p>
		</div>
	{:else if vulnerabilities.length === 0}
		<div class="card text-center py-12">
			<p class="text-gray-500">Vulnerability not found</p>
		</div>
	{:else}
		<!-- CVE Summary -->
		<div class="card">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Vulnerability Details</h2>
			<dl class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div>
					<dt class="text-sm font-medium text-gray-500">CVE ID</dt>
					<dd class="mt-1 text-sm text-gray-900">{cveInfo?.cve_id}</dd>
				</div>
				<div>
					<dt class="text-sm font-medium text-gray-500">Severity</dt>
					<dd class="mt-1">
						<SeverityBadge severity={cveInfo?.severity || 'Unknown'} />
					</dd>
				</div>
				<div class="md:col-span-2">
					<dt class="text-sm font-medium text-gray-500">Description</dt>
					<dd class="mt-1 text-sm text-gray-900">
						{cveInfo?.description || 'No description available'}
					</dd>
				</div>
				{#if cveInfo?.url}
					<div class="md:col-span-2">
						<dt class="text-sm font-medium text-gray-500">Reference</dt>
						<dd class="mt-1 text-sm">
							<a href={cveInfo.url} target="_blank" rel="noopener noreferrer" class="text-blue-600 hover:underline">
								{cveInfo.url}
							</a>
						</dd>
					</div>
				{/if}
			</dl>
		</div>

		<!-- Affected Packages -->
		<div class="card">
			<h2 class="text-xl font-bold text-gray-900 mb-4">
				Affected Packages ({vulnerabilities.length})
			</h2>
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Package</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Version</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Fix Version</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">First Detected</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each vulnerabilities as vuln}
							<tr class="hover:bg-gray-50">
								<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
									{vuln.package_name}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
									{vuln.package_version}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
									{vuln.package_type || 'N/A'}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<StatusBadge status={vuln.status} />
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
									{vuln.fix_version || 'N/A'}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
									{formatDate(vuln.first_detected_at)}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<button
										on:click={() => openUpdateModal(vuln)}
										class="text-blue-600 hover:text-blue-800"
									>
										Update
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

<!-- Update Modal -->
{#if showUpdateModal && selectedVuln}
	<div class="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
		<div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
			<div class="px-6 py-4 border-b border-gray-200">
				<h3 class="text-lg font-semibold text-gray-900">Update Vulnerability</h3>
				<p class="text-sm text-gray-600 mt-1">
					{selectedVuln.cve_id} - {selectedVuln.package_name} {selectedVuln.package_version}
				</p>
			</div>

			<div class="px-6 py-4 space-y-4">
				<div>
					<label for="status" class="block text-sm font-medium text-gray-700">Status</label>
					<select
						id="status"
						bind:value={updateStatus}
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
					>
						<option value="active">Active</option>
						<option value="fixed">Fixed</option>
						<option value="ignored">Ignored</option>
						<option value="accepted">Accepted</option>
					</select>
				</div>

				<div>
					<label for="notes" class="block text-sm font-medium text-gray-700">Notes</label>
					<textarea
						id="notes"
						bind:value={updateNotes}
						rows="3"
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						placeholder="Add notes about this vulnerability..."
					></textarea>
				</div>
			</div>

			<div class="px-6 py-4 bg-gray-50 flex justify-end space-x-3 rounded-b-lg">
				<button
					on:click={closeUpdateModal}
					disabled={updating}
					class="btn btn-secondary"
				>
					Cancel
				</button>
				<button
					on:click={updateVulnerability}
					disabled={updating}
					class="btn btn-primary"
				>
					{updating ? 'Updating...' : 'Update'}
				</button>
			</div>
		</div>
	</div>
{/if}
