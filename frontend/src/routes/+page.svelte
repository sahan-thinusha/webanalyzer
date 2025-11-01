<script lang="ts">
	import { onMount } from 'svelte';
	import { writable } from 'svelte/store';

	const url = writable('');
	const result = writable<any>(null);
	const loading = writable(false);
	const error = writable('');

	async function analyze() {
		error.set('');
		result.set(null);

		const targetURL = $url.trim();
		if (!targetURL) {
			error.set('Please enter a URL to analyze');
			return;
		}

		loading.set(true);
		try {
			const apiUrl = `http://localhost:8080/webanalyzer/api/v1/analyze?url=${encodeURIComponent(targetURL)}`;
			const response = await fetch(apiUrl);

			if (response.status !== 200) {
				throw new Error(`HTTP Error: ${response.status}`);
			}

			const json = await response.json();

			result.set(json.data);
		} catch (err) {
			error.set(err.message || 'Failed to fetch analysis');
		} finally {
			loading.set(false);
		}
	}
</script>

<div class="mx-auto max-w-3xl p-8">
	<h1 class="mb-6 text-2xl font-semibold">Web Analyzer</h1>

	<div class="mb-6 flex gap-2">
		<input
			class="flex-1 rounded-lg border px-4 py-2"
			type="url"
			placeholder="Enter website URL (e.g., https://facebook.com)"
			bind:value={$url}
		/>
		<button
			class="rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
			on:click={analyze}
			disabled={$loading}
		>
			{$loading ? 'Analyzing...' : 'Analyze'}
		</button>
	</div>

	{#if $error}
		<div class="mb-4 font-medium text-red-600">{$error}</div>
	{/if}

	{#if $result}
		<div class="rounded-lg border bg-gray-50 p-4">
			<h2 class="mb-3 text-xl font-semibold">Analysis Result</h2>

			<div><strong>HTML Version:</strong> {$result.html_version}</div>
			<div><strong>Page Title:</strong> {$result.page_title}</div>

			<h3 class="mt-4 font-semibold">Headings</h3>
			<ul class="ml-4 list-disc">
				<li>H1: {$result.heading_counts.h1}</li>
				<li>H2: {$result.heading_counts.h2}</li>
				<li>H3: {$result.heading_counts.h3}</li>
				<li>H4: {$result.heading_counts.h4}</li>
				<li>H5: {$result.heading_counts.h5}</li>
				<li>H6: {$result.heading_counts.h6}</li>
			</ul>

			<h3 class="mt-4 font-semibold">Links</h3>
			<ul class="ml-4 list-disc">
				<li>Internal Links: {$result.internal_link_count}</li>
				<li>External Links: {$result.external_link_count}</li>
				<li>Inaccessible Links: {$result.inaccessible_link_count}</li>
			</ul>

			<div class="mt-4">
				<strong>Has Login Form:</strong>
				{$result.has_login_form ? '✅ Yes' : '❌ No'}
			</div>
		</div>
	{/if}
</div>

<style>
	input:disabled,
	button:disabled {
		opacity: 0.7;
		cursor: not-allowed;
	}
</style>
