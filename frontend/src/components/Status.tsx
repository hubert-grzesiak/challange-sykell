import React from 'react';
import { Chip } from '@mui/material';
import { CheckCircle, Error, HourglassEmpty } from '@mui/icons-material';

interface StatusProps {
  status: 'queued' | 'running' | 'done' | 'error' | 'stopped';
}

const Status: React.FC<StatusProps> = ({ status }) => {
  const getStatusChip = () => {
    switch (status) {
      case 'queued':
        return <Chip icon={<HourglassEmpty />} label="Queued" color="warning" />;
      case 'running':
        return <Chip icon={<HourglassEmpty />} label="Running" color="info" />;
      case 'done':
        return <Chip icon={<CheckCircle />} label="Done" color="success" />;
      case 'error':
        return <Chip icon={<Error />} label="Error" color="error" />;
      case 'stopped':
        return <Chip icon={<Error />} label="Stopped" color="default" />;
      default:
        return null;
    }
  };

  return getStatusChip();
};

export default Status;