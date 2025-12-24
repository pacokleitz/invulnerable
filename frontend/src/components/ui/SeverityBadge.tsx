import { FC } from 'react';

interface SeverityBadgeProps {
	severity: string;
}

export const SeverityBadge: FC<SeverityBadgeProps> = ({ severity }) => {
	const badgeClass =
		{
			Critical: 'badge-critical',
			High: 'badge-high',
			Medium: 'badge-medium',
			Low: 'badge-low'
		}[severity] || 'badge-low';

	return <span className={`badge ${badgeClass}`}>{severity}</span>;
};
