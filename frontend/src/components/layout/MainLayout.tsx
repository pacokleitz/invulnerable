import { FC, useEffect } from 'react';
import { NavLink, Outlet, useSearchParams } from 'react-router-dom';
import { useStore } from '../../store';

export const MainLayout: FC = () => {
	const { user, loadUser } = useStore();
	const [searchParams] = useSearchParams();

	useEffect(() => {
		loadUser();
	}, [loadUser]);

	const linkClass = ({ isActive }: { isActive: boolean }) =>
		`inline-flex items-center px-1 pt-1 text-sm font-medium ${
			isActive ? 'text-blue-600' : 'text-gray-500 hover:text-blue-600'
		}`;

	// Preserve image and severity filters across pages
	const getNavLinkWithFilters = (path: string) => {
		const imageFilter = searchParams.get('image');
		const severityFilter = searchParams.get('severity');
		const params = new URLSearchParams();

		if (imageFilter) params.set('image', imageFilter);
		if (severityFilter) params.set('severity', severityFilter);

		const queryString = params.toString();
		return queryString ? `${path}?${queryString}` : path;
	};

	return (
		<div className="min-h-screen bg-gray-50">
			<nav className="bg-white shadow-sm border-b border-gray-200" role="navigation" aria-label="Main navigation">
				<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
					<div className="flex justify-between h-16">
						<div className="flex">
							<div className="flex-shrink-0 flex items-center">
								<NavLink to={getNavLinkWithFilters('/')} className="text-xl font-bold text-blue-600" aria-label="Home">
									Invulnerable
								</NavLink>
							</div>
							<div className="ml-6 flex space-x-8" role="menubar">
								<NavLink to={getNavLinkWithFilters('/')} className={linkClass} end role="menuitem">
									Dashboard
								</NavLink>
								<NavLink to={getNavLinkWithFilters('/scans')} className={linkClass} role="menuitem">
									Scans
								</NavLink>
								<NavLink to={getNavLinkWithFilters('/vulnerabilities')} className={linkClass} role="menuitem">
									Vulnerabilities
								</NavLink>
								<NavLink to={getNavLinkWithFilters('/images')} className={linkClass} role="menuitem">
									Images
								</NavLink>
							</div>
						</div>
						<div className="flex items-center space-x-4">
							{user && (
								<>
									<span className="text-sm text-gray-700">{user.email}</span>
									<a
										href="/oauth2/sign_out"
										className="text-sm text-gray-500 hover:text-gray-700"
									>
										Logout
									</a>
								</>
							)}
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
