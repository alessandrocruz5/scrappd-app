import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../providers/auth_provider.dart';
import '../items/create_item_screen.dart';
import '../items/items_gallery_screen.dart';
import '../pages/page_editor_screen.dart';

class MainShell extends StatefulWidget {
  const MainShell({super.key, this.initialIndex = 1});

  final int initialIndex;

  static MainShellState? of(BuildContext context) {
    return context.findAncestorStateOfType<MainShellState>();
  }

  @override
  State<MainShell> createState() => MainShellState();
}

class MainShellState extends State<MainShell> {
  late int _index;

  @override
  void initState() {
    super.initState();
    _index = widget.initialIndex;
  }

  void setIndex(int value) {
    if (_index == value) return;
    setState(() {
      _index = value;
    });
  }

  @override
  Widget build(BuildContext context) {
    final pages = [
      const CreateItemScreen(),
      const ItemsGalleryScreen(),
      const PageEditorScreen(),
    ];

    return Scaffold(
      appBar: AppBar(
        title: Text(
          _index == 0
              ? 'Create'
              : _index == 1
              ? 'Gallery'
              : 'Pages',
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () => context.read<AuthProvider>().logout(),
            tooltip: 'Log out',
          ),
        ],
      ),
      body: pages[_index],
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (value) {
          setIndex(value);
        },
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.add_a_photo_outlined),
            selectedIcon: Icon(Icons.add_a_photo),
            label: 'Create',
          ),
          NavigationDestination(
            icon: Icon(Icons.grid_view_outlined),
            selectedIcon: Icon(Icons.grid_view),
            label: 'Gallery',
          ),
          NavigationDestination(
            icon: Icon(Icons.auto_awesome_mosaic_outlined),
            selectedIcon: Icon(Icons.auto_awesome_mosaic),
            label: 'Pages',
          ),
        ],
      ),
    );
  }
}
