import '../models/table_info.dart';

class LobbyRepository {
  Future<List<TableInfo>> fetchTables(String type) async {
    // TODO: call REST API
    await Future.delayed(const Duration(milliseconds: 500));
    return [
      const TableInfo(
        id: 't1',
        name: '经典六人桌',
        stakes: '5/10',
        maxPlayers: 6,
        currentPlayers: 5,
        limit: 5000,
        isFull: false,
      ),
      const TableInfo(
        id: 't2',
        name: '快速九人桌',
        stakes: '10/20',
        maxPlayers: 9,
        currentPlayers: 8,
        limit: 5000,
        isFull: false,
      ),
      const TableInfo(
        id: 't3',
        name: '高手十人桌',
        stakes: '50/100',
        maxPlayers: 10,
        currentPlayers: 10,
        limit: 5000,
        isFull: true,
      ),
    ];
  }
}
