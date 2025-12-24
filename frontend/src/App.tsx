import { FC } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ErrorBoundary } from './components/ErrorBoundary';
import { MainLayout } from './components/layout/MainLayout';
import { Dashboard } from './components/pages/Dashboard';
import { ScansList } from './components/pages/ScansList';
import { ScanDetails } from './components/pages/ScanDetails';
import { ScanSBOM } from './components/pages/ScanSBOM';
import { ScanDiff } from './components/pages/ScanDiff';
import { VulnerabilitiesList } from './components/pages/VulnerabilitiesList';
import { CVEDetails } from './components/pages/CVEDetails';
import { ImagesList } from './components/pages/ImagesList';
import { ImageHistory } from './components/pages/ImageHistory';

export const App: FC = () => {
	return (
		<ErrorBoundary>
			<BrowserRouter>
				<Routes>
					<Route path="/" element={<MainLayout />}>
						<Route index element={<Dashboard />} />
						<Route path="scans" element={<ScansList />} />
						<Route path="scans/:id" element={<ScanDetails />} />
						<Route path="scans/:id/sbom" element={<ScanSBOM />} />
						<Route path="scans/:id/diff" element={<ScanDiff />} />
						<Route path="vulnerabilities" element={<VulnerabilitiesList />} />
						<Route path="vulnerabilities/:cve" element={<CVEDetails />} />
						<Route path="images" element={<ImagesList />} />
						<Route path="images/:id" element={<ImageHistory />} />
					</Route>
				</Routes>
			</BrowserRouter>
		</ErrorBoundary>
	);
};
