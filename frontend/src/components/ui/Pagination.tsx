import { FC } from 'react';

interface PaginationProps {
	currentPage: number;
	totalItems: number;
	itemsPerPage: number;
	onPageChange: (page: number) => void;
	maxVisiblePages?: number;
}

export const Pagination: FC<PaginationProps> = ({
	currentPage,
	totalItems,
	itemsPerPage,
	onPageChange,
	maxVisiblePages = 5,
}) => {
	const totalPages = Math.ceil(totalItems / itemsPerPage);

	// Don't show pagination if there's only one page or no items
	if (totalPages <= 1) {
		return null;
	}

	// Calculate page range to show
	const getPageRange = () => {
		const pages: (number | string)[] = [];

		if (totalPages <= maxVisiblePages) {
			// Show all pages if total is less than max visible
			for (let i = 1; i <= totalPages; i++) {
				pages.push(i);
			}
		} else {
			// Always show first page
			pages.push(1);

			// Calculate start and end of visible range
			let start = Math.max(2, currentPage - Math.floor(maxVisiblePages / 2));
			let end = Math.min(totalPages - 1, start + maxVisiblePages - 3);

			// Adjust start if we're near the end
			if (end === totalPages - 1) {
				start = Math.max(2, end - maxVisiblePages + 3);
			}

			// Add ellipsis after first page if needed
			if (start > 2) {
				pages.push('...');
			}

			// Add middle pages
			for (let i = start; i <= end; i++) {
				pages.push(i);
			}

			// Add ellipsis before last page if needed
			if (end < totalPages - 1) {
				pages.push('...');
			}

			// Always show last page
			pages.push(totalPages);
		}

		return pages;
	};

	const pages = getPageRange();

	return (
		<div className="flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6">
			<div className="flex flex-1 justify-between sm:hidden">
				{/* Mobile pagination */}
				<button
					onClick={() => onPageChange(currentPage - 1)}
					disabled={currentPage === 1}
					className="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
				>
					Previous
				</button>
				<button
					onClick={() => onPageChange(currentPage + 1)}
					disabled={currentPage === totalPages}
					className="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
				>
					Next
				</button>
			</div>
			<div className="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
				<div>
					<p className="text-sm text-gray-700">
						Showing{' '}
						<span className="font-medium">{Math.min((currentPage - 1) * itemsPerPage + 1, totalItems)}</span>{' '}
						to{' '}
						<span className="font-medium">{Math.min(currentPage * itemsPerPage, totalItems)}</span>{' '}
						of{' '}
						<span className="font-medium">{totalItems}</span>{' '}
						results
					</p>
				</div>
				<div>
					<nav className="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
						{/* Previous button */}
						<button
							onClick={() => onPageChange(currentPage - 1)}
							disabled={currentPage === 1}
							className="relative inline-flex items-center rounded-l-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed"
						>
							<span className="sr-only">Previous</span>
							<svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
								<path fillRule="evenodd" d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z" clipRule="evenodd" />
							</svg>
						</button>

						{/* Page numbers */}
						{pages.map((page, idx) => {
							if (page === '...') {
								return (
									<span
										key={`ellipsis-${idx}`}
										className="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-gray-700 ring-1 ring-inset ring-gray-300"
									>
										...
									</span>
								);
							}

							const pageNum = page as number;
							const isCurrent = pageNum === currentPage;

							return (
								<button
									key={pageNum}
									onClick={() => onPageChange(pageNum)}
									className={`relative inline-flex items-center px-4 py-2 text-sm font-semibold ${
										isCurrent
											? 'z-10 bg-blue-600 text-white focus:z-20 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600'
											: 'text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0'
									}`}
								>
									{pageNum}
								</button>
							);
						})}

						{/* Next button */}
						<button
							onClick={() => onPageChange(currentPage + 1)}
							disabled={currentPage === totalPages}
							className="relative inline-flex items-center rounded-r-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed"
						>
							<span className="sr-only">Next</span>
							<svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
								<path fillRule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clipRule="evenodd" />
							</svg>
						</button>
					</nav>
				</div>
			</div>
		</div>
	);
};
