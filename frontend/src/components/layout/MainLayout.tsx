import { FC } from 'react';
import { NavLink, Outlet } from 'react-router-dom';

export const MainLayout: FC = () => {
	const linkClass = ({ isActive }: { isActive: boolean }) =>
		`inline-flex items-center px-1 pt-1 text-sm font-medium ${
			isActive ? 'text-blue-600' : 'text-gray-500 hover:text-blue-600'
		}`;

	return (
		<div className="min-h-screen bg-gray-50">
			<nav className="bg-white shadow-sm border-b border-gray-200" role="navigation" aria-label="Main navigation">
				<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
					<div className="flex justify-between h-16">
						<div className="flex">
							<div className="flex-shrink-0 flex items-center">
								<NavLink to="/" className="text-xl font-bold text-blue-600" aria-label="Home">
									Invulnerable
								</NavLink>
							</div>
							<div className="ml-6 flex space-x-8" role="menubar">
								<NavLink to="/" className={linkClass} end role="menuitem">
									Dashboard
								</NavLink>
								<NavLink to="/scans" className={linkClass} role="menuitem">
									Scans
								</NavLink>
								<NavLink to="/vulnerabilities" className={linkClass} role="menuitem">
									Vulnerabilities
								</NavLink>
								<NavLink to="/images" className={linkClass} role="menuitem">
									Images
								</NavLink>
							</div>
						</div>
					</div>
				</div>
			</nav>

			<main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8" role="main">
				<Outlet />
			</main>
		</div>
	);
};
